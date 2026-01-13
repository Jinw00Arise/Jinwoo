package game

type Job int16

const (
	// Beginners
	JobBeginner     Job = 0
	JobNoblesse     Job = 1000
	JobAranBeginner Job = 2000
	JobEvanBeginner Job = 2001
	JobCitizen      Job = 3000

	// Warriors
	JobSwordsman    Job = 100
	JobFighter      Job = 110
	JobCrusader     Job = 111
	JobHero         Job = 112
	JobPage         Job = 120
	JobWhiteKnight  Job = 121
	JobPaladin      Job = 122
	JobSpearman     Job = 130
	JobDragonKnight Job = 131
	JobDarkKnight   Job = 132

	// Magicians
	JobMagician   Job = 200
	JobFPWizard   Job = 210
	JobFPMage     Job = 211
	JobFPArchMage Job = 212
	JobILWizard   Job = 220
	JobILMage     Job = 221
	JobILArchMage Job = 222
	JobCleric     Job = 230
	JobPriest     Job = 231
	JobBishop     Job = 232

	// Archers
	JobArcher      Job = 300
	JobHunter      Job = 310
	JobRanger      Job = 311
	JobBowmaster   Job = 312
	JobCrossbowman Job = 320
	JobSniper      Job = 321
	JobMarksman    Job = 322

	// Thieves
	JobThief           Job = 400
	JobAssassin        Job = 410
	JobHermit          Job = 411
	JobNightLord       Job = 412
	JobBandit          Job = 420
	JobChiefBandit     Job = 421
	JobShadower        Job = 422
	JobBladeRecruit    Job = 430
	JobBladeAcolyte    Job = 431
	JobBladeSpecialist Job = 432
	JobBladeLord       Job = 433
	JobBladeMaster     Job = 434

	// Pirates
	JobPirate     Job = 500
	JobBrawler    Job = 510
	JobMarauder   Job = 511
	JobBuccaneer  Job = 512
	JobGunslinger Job = 520
	JobOutlaw     Job = 521
	JobCorsair    Job = 522

	// Cygnus Knights
	JobDawnWarrior1    Job = 1100
	JobDawnWarrior2    Job = 1110
	JobDawnWarrior3    Job = 1111
	JobBlazeWizard1    Job = 1200
	JobBlazeWizard2    Job = 1210
	JobBlazeWizard3    Job = 1211
	JobWindArcher1     Job = 1300
	JobWindArcher2     Job = 1310
	JobWindArcher3     Job = 1311
	JobNightWalker1    Job = 1400
	JobNightWalker2    Job = 1410
	JobNightWalker3    Job = 1411
	JobThunderBreaker1 Job = 1500
	JobThunderBreaker2 Job = 1510
	JobThunderBreaker3 Job = 1511

	// Aran
	JobAran1 Job = 2100
	JobAran2 Job = 2110
	JobAran3 Job = 2111
	JobAran4 Job = 2112

	// Evan
	JobEvan1  Job = 2200
	JobEvan2  Job = 2210
	JobEvan3  Job = 2211
	JobEvan4  Job = 2212
	JobEvan5  Job = 2213
	JobEvan6  Job = 2214
	JobEvan7  Job = 2215
	JobEvan8  Job = 2216
	JobEvan9  Job = 2217
	JobEvan10 Job = 2218

	// Game Masters
	JobGM      Job = 500
	JobSuperGM Job = 510
)

func (j Job) String() string {
	switch j {
	case JobBeginner:
		return "Beginner"
	case JobNoblesse:
		return "Noblesse"
	case JobAranBeginner:
		return "Aran Beginner"
	case JobEvanBeginner:
		return "Evan Beginner"
	case JobCitizen:
		return "Citizen"
	// Add more as needed
	default:
		return "Unknown"
	}
}

func (j Job) IsBeginner() bool {
	return j == JobBeginner || j == JobNoblesse || j == JobAranBeginner ||
		j == JobEvanBeginner || j == JobCitizen
}
