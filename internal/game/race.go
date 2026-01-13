package game

type Race int

const (
	RaceResistance Race = 0
	RaceNormal     Race = 1 // explorer
	RaceCygnus     Race = 2
	RaceAran       Race = 3
	RaceEvan       Race = 4
)

type RaceInfo struct {
	Race Race
	Job  Job
}

var raceToJob = map[Race]Job{
	RaceResistance: JobCitizen,
	RaceNormal:     JobBeginner,
	RaceCygnus:     JobNoblesse,
	RaceAran:       JobAranBeginner,
	RaceEvan:       JobEvanBeginner,
}

func GetJobByRace(race Race) (Job, bool) {
	job, exists := raceToJob[race]
	return job, exists
}

func (r Race) GetJob() Job {
	if job, exists := raceToJob[r]; exists {
		return job
	}
	return JobBeginner // default fallback
}
