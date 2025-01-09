package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"thirdai_platform/model_bazaar/config"
	"time"
)

type KnowledgeExtractionClient struct {
	ModelClient
}

type createReportParams struct {
	Documents []config.FileInfo `json:"documents"`
}

type createReportResponse struct {
	Data struct {
		ReportId string `json:"report_id"`
	} `json:"data"`
}

func (c *KnowledgeExtractionClient) CreateReport(files []config.FileInfo) (string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	err := addFilesToMultipart(writer, files)
	if err != nil {
		return "", err
	}

	form, err := writer.CreateFormField("documents")
	if err != nil {
		return "", fmt.Errorf("error creating documents field in request: %w", err)
	}
	documents := createReportParams{Documents: files}
	for i := range documents.Documents {
		if documents.Documents[i].Options == nil {
			documents.Documents[i].Options = map[string]interface{}{}
		}
	}
	err = json.NewEncoder(form).Encode(documents)
	if err != nil {
		return "", fmt.Errorf("error serializing documents to json")
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("error closing multipart writer: %w", err)
	}

	var res createReportResponse
	err = c.Post(fmt.Sprintf("/%v/report/create", c.deploymentId())).Header("Content-Type", writer.FormDataContentType()).Body(body).Do(&res)
	if err != nil {
		return "", err
	}

	return res.Data.ReportId, nil
}

type Report struct {
	ReportId string `json:"report_id"`
	Status   string `json:"status"`
	Msg      string `json:"msg"`
	Content  struct {
		Results []struct {
			QuestionId string `json:"question_id"`
			Question   string `json:"question"`
			Answer     string `json:"answer"`
			References []struct {
				Text   string `json:"text"`
				Source string `json:"source"`
			}
		} `json:"results"`
	} `json:"content"`
}

func (c *KnowledgeExtractionClient) GetReport(reportId string) (Report, error) {
	var res wrappedData[Report]
	err := c.Get(fmt.Sprintf("/%v/report/%v", c.deploymentId(), reportId)).Do(&res)
	return res.Data, err
}

func (c *KnowledgeExtractionClient) AwaitReport(reportId string, timeout time.Duration) (Report, error) {
	check := time.Tick(2 * time.Second)
	stop := time.Tick(timeout)

	for {
		select {
		case <-check:
			report, err := c.GetReport(reportId)
			if err != nil {
				return Report{}, err
			}
			if report.Status == "complete" || report.Status == "failed" {
				return report, nil
			}
		case <-stop:
			return Report{}, fmt.Errorf("timeout reached before report completed")
		}
	}
}

func (c *KnowledgeExtractionClient) DeleteReport(reportId string) error {
	return c.Delete(fmt.Sprintf("/%v/report/%v", c.deploymentId(), reportId)).Do(nil)
}

func (c *KnowledgeExtractionClient) ListReports() ([]Report, error) {
	var res wrappedData[[]Report]
	err := c.Get(fmt.Sprintf("/%v/reports", c.deploymentId())).Do(&res)
	return res.Data, err
}

type Question struct {
	QuestionId   string   `json:"question_id"`
	QuestionText string   `json:"question_text"`
	Keywords     []string `json:"keywords"`
}

func (c *KnowledgeExtractionClient) ListQuestions() ([]Question, error) {
	var res wrappedData[[]Question]
	err := c.Get(fmt.Sprintf("/%v/questions", c.deploymentId())).Do(&res)
	return res.Data, err
}

func (c *KnowledgeExtractionClient) AddQuestion(question string) error {
	return c.Post(fmt.Sprintf("/%v/questions", c.deploymentId())).Param("question", question).Do(nil)
}

func (c *KnowledgeExtractionClient) DeleteQuestion(questionId string) error {
	return c.Delete(fmt.Sprintf("/%v/questions/%v", c.deploymentId(), questionId)).Do(nil)
}

func (c *KnowledgeExtractionClient) AddKeywords(questionId string, keywords []string) error {
	return c.Post(fmt.Sprintf("/%v/questions/%v/keywords", c.deploymentId(), questionId)).Json(keywords).Do(nil)
}
