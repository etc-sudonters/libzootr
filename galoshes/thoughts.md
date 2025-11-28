Galoshes - user query interface for table data

Basis?

SQL

Pros:
- most familiar
- very expressive
- supports graph queries since SQL:2023 via SQL/PG (property graphs)[^fn1]
- current data store is like a KV store


Cons:
- syntax is complex
- lots of features that aren't needed
- graph syntax is ...not great

```
SELECT name FROM GRAPH_TABLE (characters
    MATCH (a IS node WHERE a.name='Samwise Gamgee') <-[e IS relationship WHERE e.type='parent']- (b IS node)
    COLUMNS (b.name AS name)
);
```

Datalog?

Pros:
- relatively (to sql) simpler grammar
- existing data store maps basically 1:1 to datalog triples[^fn2]
- per query extensible via rules

Cons:
- less familiar

Notes: I found it intuitive after looking at it for fifteen minutes. 

```
; where is the kokri sword placed?
find [ $place-id $place-name ]
where [
    [ $place-id world/placement/holds $token-id ]
    [ :named $place-id $place-name]
    [ :named $oktne-id "Kokri Sword"]
]
rules [
[    [:named $id $name] [$id names $name] ]
]

; what dungeon is at SFM 
find [ $dungeonId $dungeonName ]
where [
    [ :edge $sfmLedge $nodeId ]
    [ $sfmLedge names "SFM Forest Temple Ledge" ]
    [ $nodeId world/region $dungeonId ]
    [ $dungeonId names $dungeonName ]
    [ $dungeonId world/region/kind "Dungeon" ]
]

;
find [ $dungeonId $dungeonName ]
with [ $entranceName ]
where [
   [ $entranceId names $entranceName ]
   [ :edge $entranceId $nodeId ]
   [ $nodeId world/region $dungeonId ]
   [ $dungeonId names $dungeonName ]
   [ $dungeonId world/region/kind "Dungeon" ]
]

; save a query 
$named :- find [ $id ] with [ $name ] where [ $id names $tokenName ]
$in-region :- find [ ] with [ $nodeId $regionId ] where [ $nodeId world/region $regionId ]
$in-dungeon :- find [ ] with [ $nodeId $dungeonId ] where [
    [:in-region $nodeId $dungeonId]
    [$dungeonId world/region/kind "Dungeon"]
]

$uses-token-by-name :- find [ $edgeId $ruleId ]
with [ $tokenName ]
where [
    [ :named $tokenId $tokenName ]
    [ $ruleId world/logic/rule $rule ]
    [ $edgeId world/edge/rule $ruleId ]
    [ :rule-uses-token $rule $tokenId ]
]

$storm-grotto-scrubs :- find [ $stormGrottoScrubId ]
where [
    [ $stormGrottoScrubId components/grotto-scrub 1 ] ; [ :grotto-scrub $stormGrottoScrubId ] -- can generate tag rules like this
    [ :edge _ $stormGrottoScrubId $grottoId ] 
    [ :edge $edgeId $grottoId $exit ]
    [ :uses-token-by-name "Song of Storms" $edgeId $ruleId ]
]

; place a triforce piece on storm grotto scrubs
insert [ [ $nodeId world/placement/holds $tokenId ] ]
where [
    [ :storm-grotto-scrubs $nodeId ]
    [ :named "Triforce Piece" $tokenId ]
]

```
---

```
...
    [ $the-id some/attr _ ]
    [ $the-id another/attr _ ]
      ^
      |__ need intersection of column indexes 

    map[Var][]table.ColumnId -> $the-id: { 15, 94 }
    map[Attr]table.ColumnId  -> some/attr: 94, another/attr: 15
...
```




---
https://github.com/knowsys/nemo
https://www.instantdb.com/essays/datalogjs
https://github.com/fkettelhoit/bottom-up-datalog-js
https://buttondown.com/tensegritics-curiosities/archive/writing-the-worst-datalog-ever-in-26loc/
https://buttondown.com/tensegritics-curiosities/archive/half-dumb-datalog-in-30-loc/
https://buttondown.com/tensegritics-curiosities/archive/restrained-datalog-in-39loc/
https://ajmmertens.medium.com/a-roadmap-to-entity-relationships-5b1d11ebb4eb
https://ajmmertens.medium.com/building-games-in-ecs-with-entity-relationships-657275ba2c6c
https://en.wikipedia.org/wiki/Datalog
https://en.wikipedia.org/wiki/Backtracking
https://en.wikipedia.org/wiki/Constraint_satisfaction_problem
https://en.wikipedia.org/wiki/Structured_programming
https://en.wikipedia.org/wiki/Control-flow_graph#Reducibility
https://sarabander.github.io/sicp/html/index.xhtml#SEC_Contents

https://github.com/eliben/code-for-blog/blob/main/2018/type-inference/typing.py
https://eli.thegreenplace.net/2018/type-inference/
https://medium.com/@dhruvrajvanshi/type-inference-for-beginners-part-1-3e0a5be98a4b
https://medium.com/@dhruvrajvanshi/type-inference-for-beginners-part-2-f39c33ca9513
https://www.plai.org/3/2/PLAI%20Version%203.2.2%20electronic.pdf
https://okmij.org/ftp/Computation/Computation.html#teval
https://github.com/milesbarr/hindley-milner-in-python
https://github.com/chewxy/hm
---

[^fn1]: https://www.enterprisedb.com/blog/representing-graphs-postgresql-sqlpgq

[^fn2]: https://ajmmertens.medium.com/why-it-is-time-to-start-thinking-of-games-as-databases-e7971da33ac3 flecs also thinks so
