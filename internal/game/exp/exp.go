package exp

// ExpTable contains the EXP required to level up for each level (1-200)
// Index 0 is unused, index N contains EXP needed to go from level N to N+1
var ExpTable = []int64{
	0,       // Level 0 (unused)
	15,      // Level 1 -> 2
	34,      // Level 2 -> 3
	57,      // Level 3 -> 4
	92,      // Level 4 -> 5
	135,     // Level 5 -> 6
	372,     // Level 6 -> 7
	560,     // Level 7 -> 8
	840,     // Level 8 -> 9
	1242,    // Level 9 -> 10
	1716,    // Level 10 -> 11
	2360,    // Level 11 -> 12
	3216,    // Level 12 -> 13
	4200,    // Level 13 -> 14
	5460,    // Level 14 -> 15
	7050,    // Level 15 -> 16
	8840,    // Level 16 -> 17
	11040,   // Level 17 -> 18
	13716,   // Level 18 -> 19
	16680,   // Level 19 -> 20
	20216,   // Level 20 -> 21
	24402,   // Level 21 -> 22
	29282,   // Level 22 -> 23
	34950,   // Level 23 -> 24
	41480,   // Level 24 -> 25
	48960,   // Level 25 -> 26
	57468,   // Level 26 -> 27
	67056,   // Level 27 -> 28
	77840,   // Level 28 -> 29
	89880,   // Level 29 -> 30
	103230,  // Level 30 -> 31
	117980,  // Level 31 -> 32
	134216,  // Level 32 -> 33
	152040,  // Level 33 -> 34
	171576,  // Level 34 -> 35
	192960,  // Level 35 -> 36
	216360,  // Level 36 -> 37
	241950,  // Level 37 -> 38
	269940,  // Level 38 -> 39
	300480,  // Level 39 -> 40
	333786,  // Level 40 -> 41
	370080,  // Level 41 -> 42
	409608,  // Level 42 -> 43
	452640,  // Level 43 -> 44
	499482,  // Level 44 -> 45
	550440,  // Level 45 -> 46
	605880,  // Level 46 -> 47
	666180,  // Level 47 -> 48
	731760,  // Level 48 -> 49
	803040,  // Level 49 -> 50
	880470,  // Level 50 -> 51
	964560,  // Level 51 -> 52
	1055850, // Level 52 -> 53
	1154880, // Level 53 -> 54
	1262220, // Level 54 -> 55
	1378440, // Level 55 -> 56
	1504158, // Level 56 -> 57
	1640016, // Level 57 -> 58
	1786680, // Level 58 -> 59
	1944840, // Level 59 -> 60
	2115252, // Level 60 -> 61
	2298720, // Level 61 -> 62
	2496096, // Level 62 -> 63
	2708280, // Level 63 -> 64
	2936232, // Level 64 -> 65
	3181008, // Level 65 -> 66
	3443736, // Level 66 -> 67
	3725640, // Level 67 -> 68
	4028028, // Level 68 -> 69
	4352316, // Level 69 -> 70
	4699992, // Level 70 -> 71
	5072616, // Level 71 -> 72
	5471832, // Level 72 -> 73
	5899368, // Level 73 -> 74
	6357048, // Level 74 -> 75
	6846792, // Level 75 -> 76
	7370592, // Level 76 -> 77
	7930536, // Level 77 -> 78
	8528808, // Level 78 -> 79
	9167688, // Level 79 -> 80
	9849576, // Level 80 -> 81
	10576920, // Level 81 -> 82
	11352312, // Level 82 -> 83
	12178416, // Level 83 -> 84
	13058000, // Level 84 -> 85
	13993920, // Level 85 -> 86
	14989152, // Level 86 -> 87
	16046784, // Level 87 -> 88
	17170024, // Level 88 -> 89
	18362232, // Level 89 -> 90
	19626912, // Level 90 -> 91
	20967744, // Level 91 -> 92
	22388568, // Level 92 -> 93
	23893392, // Level 93 -> 94
	25486416, // Level 94 -> 95
	27172008, // Level 95 -> 96
	28954752, // Level 96 -> 97
	30839400, // Level 97 -> 98
	32830872, // Level 98 -> 99
	34934280, // Level 99 -> 100
	37154976, // Level 100 -> 101
	39498528, // Level 101 -> 102
	41970672, // Level 102 -> 103
	44577360, // Level 103 -> 104
	47324784, // Level 104 -> 105
	50219328, // Level 105 -> 106
	53267616, // Level 106 -> 107
	56476464, // Level 107 -> 108
	59853008, // Level 108 -> 109
	63404640, // Level 109 -> 110
	67138992, // Level 110 -> 111
	71064024, // Level 111 -> 112
	75187944, // Level 112 -> 113
	79519176, // Level 113 -> 114
	84066480, // Level 114 -> 115
	88838880, // Level 115 -> 116
	93845712, // Level 116 -> 117
	99096552, // Level 117 -> 118
	104601336, // Level 118 -> 119
	110370288, // Level 119 -> 120
	116413896, // Level 120 -> 121
	122742960, // Level 121 -> 122
	129368592, // Level 122 -> 123
	136302240, // Level 123 -> 124
	143555664, // Level 124 -> 125
	151141008, // Level 125 -> 126
	159070752, // Level 126 -> 127
	167357808, // Level 127 -> 128
	176015520, // Level 128 -> 129
	185057664, // Level 129 -> 130
	194498496, // Level 130 -> 131
	204352704, // Level 131 -> 132
	214635456, // Level 132 -> 133
	225362368, // Level 133 -> 134
	236549520, // Level 134 -> 135
	248213424, // Level 135 -> 136
	260371200, // Level 136 -> 137
	273040416, // Level 137 -> 138
	286239168, // Level 138 -> 139
	299986080, // Level 139 -> 140
	314300256, // Level 140 -> 141
	329201328, // Level 141 -> 142
	344709552, // Level 142 -> 143
	360845664, // Level 143 -> 144
	377631024, // Level 144 -> 145
	395087472, // Level 145 -> 146
	413237520, // Level 146 -> 147
	432104208, // Level 147 -> 148
	451711152, // Level 148 -> 149
	472082544, // Level 149 -> 150
	493243152, // Level 150 -> 151
	515218368, // Level 151 -> 152
	538034160, // Level 152 -> 153
	561717120, // Level 153 -> 154
	586294416, // Level 154 -> 155
	611793840, // Level 155 -> 156
	638243760, // Level 156 -> 157
	665673168, // Level 157 -> 158
	694111680, // Level 158 -> 159
	723589552, // Level 159 -> 160
	754137648, // Level 160 -> 161
	785787504, // Level 161 -> 162
	818571360, // Level 162 -> 163
	852522096, // Level 163 -> 164
	887673312, // Level 164 -> 165
	924059376, // Level 165 -> 166
	961715472, // Level 166 -> 167
	1000677536, // Level 167 -> 168
	1040982384, // Level 168 -> 169
	1082668704, // Level 169 -> 170
	1125775104, // Level 170 -> 171
	1170341184, // Level 171 -> 172
	1216407552, // Level 172 -> 173
	1264015776, // Level 173 -> 174
	1313208432, // Level 174 -> 175
	1364029056, // Level 175 -> 176
	1416522192, // Level 176 -> 177
	1470733296, // Level 177 -> 178
	1526708784, // Level 178 -> 179
	1584496048, // Level 179 -> 180
	1644143520, // Level 180 -> 181
	1705700544, // Level 181 -> 182
	1769217552, // Level 182 -> 183
	1834746000, // Level 183 -> 184
	1902338304, // Level 184 -> 185
	1972047936, // Level 185 -> 186
	2043929328, // Level 186 -> 187
	2118037968, // Level 187 -> 188
	2194431312, // Level 188 -> 189
	2273167920, // Level 189 -> 190
	2354307408, // Level 190 -> 191
	2437911456, // Level 191 -> 192
	2524042800, // Level 192 -> 193
	2612765328, // Level 193 -> 194
	2704144032, // Level 194 -> 195
	2798245008, // Level 195 -> 196
	2895135504, // Level 196 -> 197
	2994883920, // Level 197 -> 198
	3097559808, // Level 198 -> 199
	3203233920, // Level 199 -> 200
}

const MaxLevel = 200

// GetExpForLevel returns the EXP required to level up from the given level
func GetExpForLevel(level int) int64 {
	if level <= 0 || level >= MaxLevel {
		return 0
	}
	if level >= len(ExpTable) {
		return ExpTable[len(ExpTable)-1]
	}
	return ExpTable[level]
}

// GetTotalExpForLevel returns total EXP accumulated to reach a specific level
func GetTotalExpForLevel(level int) int64 {
	if level <= 1 {
		return 0
	}
	var total int64
	for i := 1; i < level && i < len(ExpTable); i++ {
		total += int64(ExpTable[i])
	}
	return total
}

// CalculateLevelUp checks if a character should level up and returns the new level and remaining EXP
// Returns: newLevel, newExp, levelsGained
func CalculateLevelUp(currentLevel byte, currentExp int32) (byte, int32, int) {
	level := currentLevel
	exp := int64(currentExp)
	levelsGained := 0
	
	for level < MaxLevel {
		expNeeded := GetExpForLevel(int(level))
		if expNeeded == 0 || exp < expNeeded {
			break
		}
		
		exp -= expNeeded
		level++
		levelsGained++
	}
	
	// Cap exp at max level
	if level >= MaxLevel {
		exp = 0
	}
	
	return level, int32(exp), levelsGained
}

