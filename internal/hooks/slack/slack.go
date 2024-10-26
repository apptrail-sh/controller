package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apptrail-sh/controller/internal/model"
	"io"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
)

type SlackNotifier struct {
	WebhookURL string
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		WebhookURL: webhookURL,
	}
}

func (slack *SlackNotifier) Notify(ctx context.Context, workload model.WorkloadUpdate) error {
	log := ctrl.LoggerFrom(ctx)
	httpClient := &http.Client{}

	message := "Workload version released:\n"
	message += "```"
	message += "Kind: " + workload.Kind + "\n"
	message += "Name: " + workload.Name + "\n"
	message += "Namespace: " + workload.Namespace + "\n"
	message += "Previous Version: " + workload.PreviousVersion + "\n"
	message += "Current Version: " + workload.CurrentVersion + "\n"
	message += "```"

	type SlackMessage struct {
		Text string `json:"text"`
	}
	slackMessage := SlackMessage{
		Text: message,
	}

	jsonData, err := json.Marshal(slackMessage)
	if err != nil {
		log.Error(err, "failed to marshal slack message")
		return fmt.Errorf("failed to marshal slack message. %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", slack.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error(err, "failed to create slack request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error(err, "failed to send slack message.")
		return err
	}
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err, "failed to read response body.")
			return err
		}
		msg := fmt.Sprintf("failed to send slack request. %v. Body: %v", resp.Status, body)
		errResp := errors.New(msg)
		log.Error(errResp, msg)
		return errResp
	}
	resp.Body.Close()
	return nil
}
