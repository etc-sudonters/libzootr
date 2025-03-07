package main

import (
	"sudonters/libzootr/components"
	Q "sudonters/libzootr/galoshes/query"
	"sudonters/libzootr/galoshes/script"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/internal/query"
	"sudonters/libzootr/internal/table"
	"sudonters/libzootr/internal/table/columns"
	"sudonters/libzootr/zecs"
)

var names = []string{
	"Kokri Sword",
	"Mirror Shield",
	"Dins Fire",
	"Hover Boots",
	"Iron Boots",
	"Slingshot",
	"Bombs",
	"Bow",
	"Fire Arrows",
	"Boomerang",
	"Bombchus",
	"Progressive Hookshot",
	"Ice Arrows",
	"Faroes Wind",
	"Rutos Letter",
	"Bottle",
	"Megaton Hammer",
	"Light Arrows",
	"Lens of Truth",
	"Zeldas Lullabye",
	"Eponas Song",
	"Sarias Song",
	"Suns Song",
	"Song of Time",
	"Song of Storms",
	"Minuet of Forest",
	"Bolero of Fire",
	"Serenade of Water",
	"Nocturne of Shadows",
	"Requiem of Spirit",
	"Prelude of Light",
	"Progressive Strength",
	"Kokri Emerald",
	"Goron Ruby",
	"Zora Sapphire",
	"Forest Medallion",
	"Fire Medallion",
	"Water Medallion",
	"Shadow Medallion",
	"Spirit Medallion",
	"Light Medallion",
}

func sizedslice[T zecs.Value](attr string, size uint32) zecs.DDL {
	return func() *table.ColumnBuilder {
		return columns.SizedSliceColumn[T](attr, size)
	}
}

func main() {
	ddl := []zecs.DDL{
		sizedslice[components.Name]("root/name", 64),
	}

	ocm, err := zecs.New()
	internal.PanicOnError(err)
	internal.PanicOnError(zecs.Apply(&ocm, ddl))
	eng := ocm.Engine()
	tbl, err := query.ExtractTable(eng)
	internal.PanicOnError(err)

	for _, name := range names {
		row := tbl.InsertRow()
		internal.PanicOnError(tbl.SetValue(row, 1, components.Name(name)))
	}

	env := script.NewAstEnv()
	loadAttrsToAstEnv(tbl.AttrTable(), env)
	q := `find [ $id ] 
    where [
        [ $id root/name "Kokri Sword" ]
    ]
`
	find, err := script.ParseScript(q, env)
	internal.PanicOnError(err)
	result, err := Q.RunFindQuery(Q.BuildFind(find.(*script.FindNode)), tbl, nil)
	internal.PanicOnError(err)
	_ = result
}

func loadAttrsToAstEnv(attrs map[string]table.ColumnMetadata, env *script.AstEnv) {
	for name, meta := range attrs {
		var ty script.Type
		switch meta.Kind() {
		case table.ColumnInt, table.ColumnUint, table.ColumnFloat:
			ty = script.TypeNumber{}
		case table.ColumnBoolean:
			ty = script.TypeBool{}
		case table.ColumnString:
			ty = script.TypeString{}
		default:
			// don't map other kinds into query engine
			continue
		}

		env.AddAttr(name, meta.Id(), ty)
	}
}
