package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/krew-release-bot/pkg/actions"
	"github.com/rajatjindal/krew-release-bot/krew-release-bot/pkg/helpers"
	"github.com/rajatjindal/krew-release-bot/krew-release-bot/pkg/krew"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func main() {
	var err error
	realAction, err = initCredentials()
	if err != nil {
		logrus.Fatalf("failed to initialize credentials. error: %v", err)
	}

	logrus.Infof("user: %s, name: %q", realAction.TokenUserHandle, realAction.TokenUsername)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8082),
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/", Handle)
	logrus.Fatal(s.ListenAndServe())
}

const credentialsFile = "/var/openfaas/secrets/krew-release-bot.yaml"

var realAction actions.RealAction

func initCredentials() (actions.RealAction, error) {
	r, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return actions.RealAction{}, fmt.Errorf("failed to read credentials file with err: %s", err.Error())
	}

	t := actions.RealAction{}
	err = yaml.Unmarshal(r, &t)
	if err != nil {
		return actions.RealAction{}, err
	}

	return t, nil
}

//Handle handles the function call to function
func Handle(w http.ResponseWriter, r *http.Request) {
	body, ok := isValidSignature(r, realAction.WebhookSecret)

	if !ok {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
		return
	}

	t := github.WebHookType(r)
	if t == "" {
		logrus.Error("header 'X-GitHub-Event' not found. cannot handle this request")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("header 'X-GitHub-Event' not found."))
		return
	}

	logrus.Tracef("%s", string(body))

	if t != "release" {
		logrus.Error("unsupported event type")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unsupported event type."))
		return
	}

	e, err := github.ParseWebHook(t, body)
	if err != nil {
		logrus.Error("failed to parsepayload. error: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to parse payload."))
		return
	}

	event, ok := e.(*github.ReleaseEvent)
	if !ok {
		logrus.Error("not a release event")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("not a release event"))
		return
	}

	if event.GetAction() != "published" {
		logrus.Error("action is not published.")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("action is not published."))
		return
	}

	actionData, err := realAction.GetActionData(event)
	if err != nil {
		logrus.Error("failed to get actionData. error: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to get actionData."))
		return
	}

	if actionData.ReleaseInfo.GetPrerelease() {
		logrus.Infof("%s is a pre-release. not opening the PR", actionData.ReleaseInfo.GetTagName())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("its a prerelease."))
		return
	}

	tempdir, err := ioutil.TempDir("", "krew-index-")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to create tempdir."))
		return
	}

	logrus.Infof("will operate in tempdir %s", tempdir)
	repo, err := helpers.CloneRepos(actionData, tempdir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to clone the repo."))
		return
	}

	//https://raw.githubusercontent.com/rajatjindal/kubectl-modify-secret/master/.krew.yaml
	templateFileURI := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/.krew.yaml", actionData.RepoOwner, actionData.Repo)
	actualFile := filepath.Join(tempdir, "plugins", krew.PluginFileName(actionData.Inputs.PluginName))

	logrus.Info("validating ownership")
	err = krew.ValidateOwnership(actualFile, actionData.RepoOwner)
	if err != nil {
		logrus.Errorf("failed when validating ownership with error: %s", err.Error())
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("could not verify ownership of plugin."))
		return
	}

	logrus.Info("update plugin manifest with latest release info")

	err = krew.UpdatePluginManifest(templateFileURI, actualFile, actionData.ReleaseInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = krew.ValidatePlugin(actionData.Inputs.PluginName, actualFile)
	logrus.Infof("pushing changes to branch %s", actionData.ReleaseInfo.GetTagName())
	commit := helpers.Commit{
		Msg:        fmt.Sprintf("new version %s of %s", actionData.ReleaseInfo.GetTagName(), actionData.Inputs.PluginName),
		RemoteName: helpers.OriginNameLocal,
	}
	err = helpers.AddCommitAndPush(repo, commit, actionData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	logrus.Info("submitting the pr")
	pr, err := helpers.SubmitPR(actionData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pr))
}

func isValidSignature(r *http.Request, key string) ([]byte, bool) {
	// Assuming a non-empty header
	gotHash := strings.SplitN(r.Header.Get("X-Hub-Signature"), "=", 2)
	if gotHash[0] != "sha1" {
		return nil, false
	}

	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Infof("Cannot read the request body: %s\n", err)
		return nil, false
	}

	hash := hmac.New(sha1.New, []byte(key))
	if _, err := hash.Write(b); err != nil {
		logrus.Infof("Cannot compute the HMAC for request: %s\n", err)
		return nil, false
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	logrus.Infof("expected hash %s", expectedHash)
	return b, gotHash[1] == expectedHash
}
