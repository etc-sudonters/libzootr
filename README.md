For funsies attempt at an alt-frontend for ootr, mostly I work on whatever I
find interesting at the moment

* https://ootrandomizer.com/
* https://github.com/ootrandomizer/OoT-Randomizer

Minimal to no dependencies is a goal. Substrate is an exception because that
exists primarily as a place for me to put pieces I'm a little _too_ focused on
AND makes sense as a reusable module. 

## Compiling and Execution

See: [IceArrowVM notes](./notes/icearrow.md)

## Storage

All of this used to be a big ol `map[reflect.Type]map[int][]interface{}` but
now `table` forms the heart of this system.

`table` is columnar-esque storage system. The table is divided into
independent columns that store components. A component is essentially any type
that is used to describe a row. A row is a collection of column entries joined
by a common rowid. There is no fixed schema for a row, rather projections are
assembled as needed. However, this means a row can only possess a component
either zero or one times.

There are several options for storing columns:
- `columns.Bit`: produces a singleton value if the rowid is present in its bitset
- `columns.Hashmap`: Stores components in a `map[table.RowId]table.Value`
- `columns.Slice`: Stores components in a `[]table.Value` indexed by
  `table.RowId`[^fn1]

Other columns backed by a sparse sets or trees are possible but unimplemented
currently.

A rudimentary indexing system is present to assist finding components with
specific characteristics. This falls back to a column scan and typically
`reflect.DeepEqual` if an index isn't present on a scanned column. Every column
effectively has a bitset index tracking membership.

`table` provides some abstraction over the table, primarily it
provides an interface for gathering columns from the table and iterating over
the matching rows. For a row to match it must:

- be present in every column named by `Query.Exists` or `Query.Load`
- not be present in every column named by `Query.NotExists`

Rows are not matched by any particular property of its column value -- that is
handled by the similar `Engine.Lookup` method. Rather all that matters is if a
column has a value for a row or not.


### Pending

Investigate using "archetypes":

```go
type SongLocation struct {
    Name     components.Name
    Song     components.Song
    Location components.Location
}
var arch SongLocation

q := engine.CreateQuery()
q.LoadArchectype(arch) // not this but the idea
rows, _ := engine.Retrieve(q)

for rowid, values := range rows.All {
    (&arch).Init(rowid, values)
    fmt.Printl("%+v", arch)
    // SongLocation{Name: components.Name("first"), Song: components.Song{}, Location: components.Location{}}
}
```



## Logic

The necessary files can be dumped from a local copy of OOTR's source code via
the `dump-zootr.py` helper. This will copy over the logic files and dump item
and location representations to json files. Most of the loading of these files
is the responsibility of the calling application, but `internal/rules` handles
transforming the logic json files into bytecode. 

The logic files describe connections between game world locations -- edge rules
-- using a (subset) of Python. `rules/parser` produces an AST from the edge
rule or helper passed. `rules/runtime` accepts this AST and produces chunks of
bytecode, constants and names. `rules/runtime` also provides a VM to execute
this chunk. The VM/bytecode is currently a mostly 1:1 mapping from the AST
leading to awkwardness in decisions like not supporting `in` or subscripting
in the VM. The VM supports calling both the compiled OOTR helper functions
and Go.

OOTR takes advantage that it is written in Python and uses the stdlib provided
tools for parsing, transforming and compiling the raw edge rules into
callables. The primary transformations applied is extremely aggressive in
lining and compile time evaluation -- settings are replaced with their actual
values, helpers and "token literals" are treated closer to macros, where
possible the parser eliminates impossible branches with `can_use`[^fn7] being
noted particularly. 

### at/here

There also exists two "macros", `here` and `at`, that create more rules AND
collectibles that serve as proof the player is able to reach some arbitrary
location. The way this was explained to me is it is perfectly valid for the
placement to require the player to arrive at a location as one age and perform
an action -- say destroying a mudwall with the megaton hammer as an adult --
and then return as the opposite age to finish the task. If the placement engine
can reach that token then it has proof it can reach a specific location without
having to an expensive graph traversal.

```python
# Dodongos Cavern Climb
can_use(Boomerang) or at('Dodongos Cavern Far Bridge', True)
```

Behind the scenes
- A new event is created -- it's name will be like "Dodongos Cavern Far Bridge
    Subrule 1"
- This event is linked to the specified location -- "Dodongos Cavern Far
    Bridge" in this case -- and the rule -- a literal true here -- is parsed
    and set between them.
- The parser replaces the entire `at(...)` invocation with 
    `has('Dodongos Cavern Far Bridge Subrule 1', 1)`. 


`here` operates the same way, however the parser must be aware of what location
it is parsing rules for and use this location as the target.

```python
# Dodongos Cavern Beginning
here(can_blast_or_smash or Progressive_Strength_Upgrade) or dodongos_cavern_shortcuts
```

Behind the scenes:

- A new event is created named liked "Dodongos Cavern Beginning Subrule 1"
- This event is placed at "Dodongos Cavern Beginning" with the rule connecting
    them
- The parser replaces the entire `here(...)` with `has('Dodongos Cavern
    Beginning Subrule 1', 1)`


## Randomization and Placement

I recommend [Caleb Johnson's RandomizerAlgorithms][ln1][^fn6] as an example
code base and giving the attached paper a read for understanding the different
placement algorithms.

The game world is modeled by a directed graph built from the locations, exit
edge rules, and collectible edge rules.[^fn3] Edge rules dictate if a player is
_expected_ to be able to reach the destination -- either another game world
location, or some collectible. Note "expected" -- logic dictates that you need
Saria's Song to access Sacred Forest Meadow as an adult but it's possible to
also just backflip over Mido. 

These edge rules are used when placing items[^fn5]. Following from above, since
the placement engine is told "Saria's song is required to reach Adult Sacred
Forest Meadow" then either `Saria's Song` OR `Minuet of Forest` must be
accessible without anything found exclusively as an adult in Sacred Forest
Meadow or Forest Temple. For example, if `Minuet of Forest` is found at the
Windmill, then it's possible to place `Saria's Song` at the Adult Sacred Forest
Meadow song pickup.

A lot of logic is possibly non-obvious. In Fire Temple, Volvagia can have your
first key if the boss key is in the boss foyer and hover boots are accessible.
The boss key is for the door, and the hover boots for reaching the door without
dropping the central pillar down:

```python
# Fire Temple Near Boss -> Fire Temple Before Boss
is_adult and (fire_temple_shortcuts or logic_fire_boss_door_jump or Hover_Boots)
# Fire Temple Before Boss -> Volvagia Boss Room
Boss_Key_Fire_Temple
```

There are two other options for access, having a specific trick enabled or
having access to the temple's shortcut -- which either 

OOTR offers many ways to affect the randomization, a short example list:

* Expanding the collectible and location pools, for example including Gold
    Skulltula Tokens in the general pool which allows these tokens to appear in
    chests, as NPC rewards and allows items to appear from Gold Skulltula.
    These pools can be restricted in a similar fashion.
* Shuffling exits, entering the Kakariko Shooting Gallery might take you into
    Shadow Temple or possibly an overworld location. 
* Settings like "preplanted beans" open more locations at the start of the game
* Changing key behavior, such as enabling "keyrings" or "keysy" (no keys),
    and/or shuffling them into the general pool or a regional pool. This is
    particularly significant because keys are among the first items placed by
    OOTR.
* Adjusting the hint distribution, which doesn't affect token placement but
    hint placement is affected by token placement. If we have a "Kakariko is on
    the Path to Gohma" hint under the "default settings" BOTH copies CANNOT
    behind whatever this item is -- if it is the Kokri Sword then one copy
    _may_ be at the Deku but _both_ cannot.

I don't intend on supporting every feature that OOTR does, even for the
"default settings" there is quite a bit of that isn't captured by files
`dump-zootr.py` produces and instead exist as complicated conditional trees or
"built in" (read: Python code) that perform tasks that logic files can't (or
shouldn't at least).[^fn8]

---

[ln1]: https://github.com/etc-sudonters/RandomizerAlgorithms

[^fn1]: Note this means the length of the array is _at least_ the highest rowid
    _ever_ tracked by the column

[^fn3]: OOTR distinguishes between checks and events however I haven't found
    any benefit for this yet.


[^fn5]: There are "no logic" settings that just drop collectibles in locations
    with no assurances that completion is possible. There are still constraints
    on item placement, if "songs on song locations" is set then songs won't end
    up on dungeon rewards, chests, etc. However, your longshot that you need to
    complete the seed might be on Morpha, who requires the longshot to reach.

[^fn6]: Forked for posterity. 

[^fn7]: `can_use` is a little intimidating and for me -- a boolean impaired
    person -- hard to follow. However, the rule parser has a clever trick:
    regex the source code of helper, changing parameter names into their
    actual names or values before inlining. For example `can_use(Dins_Fire)`
    replaces the parameter name `item` with `Dins_Fire`. The parser is then
    able to determine if "item" and the compared item are the same -- are they
    the same identifier? -- and is able to drop all but the appropriate branch
    when it encounters the rule, rather than forcing runtime to step through
    each conditional.


[^fn8]: At the time I'm writing this, there is roughly 20k lines of code just
    in Python files in the base directory of the project.
