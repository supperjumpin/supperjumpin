package main

type processArtifacts struct {
	PIDPath string
	LogPath string
}

func stackProcessArtifacts(name string) processArtifacts {
	return processArtifacts{
		PIDPath: repoPath(".mage", name+".pid"),
		LogPath: repoPath(".mage", name+".log"),
	}
}
