package pkg

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/pkg/git"
)

func BuildComponentVersion(versioner projectVersioner) (string, error) {
	patchVersion, err := versioner.patchVersion()
	if err != nil {
		return "", err
	}

	metadata, err := versioner.buildMetadata()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s+%s", patchVersion, metadata), nil
}

type versioner struct {
	repoSource    string
	pathToProject string
}

func newVersioner(pathToProject string) *versioner {
	return &versioner{pathToProject: pathToProject}
}

func (v *versioner) buildMetadata() (string, error) {
	out, err := git.GetLatestCommitForPath(v.pathToProject, v.pathToProject)
	if err != nil {
		return "", errors.Wrapf(err, "failed executing git log to get build metadata in [%s]", v.pathToProject)
	}

	return out, nil
}

func (v *versioner) patchVersion() (string, error) {
	projectSource := filepath.Join(v.repoSource, v.pathToProject)
	out, err := git.DescribeTag(projectSource)
	if err != nil {
		return "", errors.Wrapf(err, "failed executing git describe to get version in [%s]", projectSource)
	}

	gitVersion := strings.Split(out, "-")
	gitTag := gitVersion[0]

	return gitTag, nil
}

type versionerWithGITTAG struct {
	versioner
	folderWithGITTAG  string
	sourcedFromBranch string
	releaseConfig     *ReleaseConfig
}

func newVersionerWithGITTAG(repoSource, pathToProject, sourcedFromBranch string, releaseConfig *ReleaseConfig) *versionerWithGITTAG {
	return &versionerWithGITTAG{
		folderWithGITTAG:  pathToProject,
		versioner:         versioner{repoSource: repoSource, pathToProject: pathToProject},
		sourcedFromBranch: sourcedFromBranch,
		releaseConfig:     releaseConfig,
	}
}

func newMultiProjectVersionerWithGITTAG(repoSource, pathToRootFolder, pathToMainProject, sourcedFromBranch string, releaseConfig *ReleaseConfig) *versionerWithGITTAG {
	return &versionerWithGITTAG{
		folderWithGITTAG:  pathToMainProject,
		versioner:         versioner{repoSource: repoSource, pathToProject: pathToRootFolder},
		sourcedFromBranch: sourcedFromBranch,
		releaseConfig:     releaseConfig,
	}
}

func (v *versionerWithGITTAG) patchVersion() (string, error) {
	return v.releaseConfig.readGitTag(v.folderWithGITTAG, v.sourcedFromBranch)
}

func (v *versionerWithGITTAG) buildMetadata() (string, error) {
	_, err := git.CheckoutRepo(v.repoSource, v.sourcedFromBranch)
	if err != nil {
		return "", errors.Cause(err)
	}

	projectSource := filepath.Join(v.repoSource, v.pathToProject)
	out, err := git.GetLatestCommitForPath(projectSource, projectSource)
	if err != nil {
		return "", errors.Wrapf(err, "failed executing git log to get build metadata in [%s]", projectSource)
	}

	return out, nil
}

type cliVersioner struct {
	versioner
	cliVersion string
}

func newCliVersioner(cliVersion, pathToProject string) *cliVersioner {
	return &cliVersioner{
		cliVersion: cliVersion,
		versioner:  versioner{pathToProject: pathToProject},
	}
}

func (v *cliVersioner) patchVersion() (string, error) {
	return v.cliVersion, nil
}
