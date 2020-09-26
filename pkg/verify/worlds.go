package verify

import "regexp"

type World struct {
	ID   int
	Name string
}

func NormalizedWorldName(worldId int) string {
	world := WorldNames[worldId]
	if worldId != world.ID {
		return ""
	}
	re := regexp.MustCompile("[^a-zA-Z]")
	re2 := regexp.MustCompile("\\[.*\\]")
	name := re.ReplaceAllString(world.Name, "")
	name = re2.ReplaceAllString(name, "")
	return name
}

var WorldNames = map[int]World{
	1001: World{1001, "Anvil Rock"},
	1002: World{1002, "Borlis Pass"},
	1003: World{1003, "Yak's Bend"},
	1004: World{1004, "Henge of Denravi"},
	1005: World{1005, "Maguuma"},
	1006: World{1006, "Sorrow's Furnace"},
	1007: World{1007, "Gate of Madness"},
	1008: World{1008, "Jade Quarry"},
	1009: World{1009, "Fort Aspenwood"},
	1010: World{1010, "Ehmry Bay"},
	1011: World{1011, "Stormbluff Isle"},
	1012: World{1012, "Darkhaven"},
	1013: World{1013, "Sanctum of Rall"},
	1014: World{1014, "Crystal Desert"},
	1015: World{1015, "Isle of Janthir"},
	1016: World{1016, "Sea of Sorrows"},
	1017: World{1017, "Tarnished Coast"},
	1018: World{1018, "Northern Shiverpeaks"},
	1019: World{1019, "Blackgate"},
	1020: World{1020, "Ferguson's Crossing"},
	1021: World{1021, "Dragonbrand"},
	1022: World{1022, "Kaineng"},
	1023: World{1023, "Devona's Rest"},
	1024: World{1024, "Eredon Terrace"},
	2001: World{2001, "Fissure of Woe"},
	2002: World{2002, "Desolation"},
	2003: World{2003, "Gandara"},
	2004: World{2004, "Blacktide"},
	2005: World{2005, "Ring of Fire"},
	2006: World{2006, "Underworld"},
	2007: World{2007, "Far Shiverpeaks"},
	2008: World{2008, "Whiteside Ridge"},
	2009: World{2009, "Ruins of Surmia"},
	2010: World{2010, "Seafarer's Rest"},
	2011: World{2011, "Vabbi"},
	2012: World{2012, "Piken Square"},
	2013: World{2013, "Aurora Glade"},
	2014: World{2014, "Gunnar's Hold"},
	2101: World{2101, "Jade Sea [FR]"},
	2102: World{2102, "Fort Ranik [FR]"},
	2103: World{2103, "Augury Rock [FR]"},
	2104: World{2104, "Vizunah Square [FR]"},
	2105: World{2105, "Arborstone [FR]"},
	2201: World{2201, "Kodash [DE]"},
	2202: World{2202, "Riverside [DE]"},
	2203: World{2203, "Elona Reach [DE]"},
	2204: World{2204, "Abaddon's Mouth [DE]"},
	2205: World{2205, "Drakkar Lake [DE]"},
	2206: World{2206, "Miller's Sound [DE]"},
	2207: World{2207, "Dzagonur [DE]"},
	2301: World{2301, "Baruch Bay [SP]"},
}
