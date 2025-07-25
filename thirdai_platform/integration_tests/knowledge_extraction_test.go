package integrationtests

import (
	"fmt"
	"strings"
	"testing"
	"thirdai_platform/client"
	"time"

	"github.com/google/uuid"
)

func checkAndProcessQuestion(t *testing.T, ke *client.KnowledgeExtractionClient, q client.Question) {
	if strings.Contains(q.QuestionText, "EPS") {
		err := ke.AddKeywords(q.QuestionId, []string{"earnings", "per", "share"})
		if err != nil {
			t.Fatal(err)
		}
	}
	if strings.Contains(q.QuestionText, "deleted") {
		err := ke.DeleteQuestion(q.QuestionId)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func checkReports(t *testing.T, ke *client.KnowledgeExtractionClient, expectedAnswers map[string]string, reuseReportId *string) string {
	var reportId string
	if reuseReportId != nil {
		reportId = *reuseReportId
	} else {
		var err error
		reportId, err = ke.CreateReport([]client.FileInfo{
			{Path: "./data/apple-10k.pdf", Location: "upload"},
		})
		if err != nil {
			t.Fatal(err)
		}

		reports, err := ke.ListReports()
		if err != nil {
			t.Fatal(err)
		}

		if len(reports) != 1 || reports[0].ReportId != reportId {
			t.Fatalf("incorrect reports returned: %v", reports)
		}
	}

	report, err := ke.AwaitReport(reportId, 200*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	questionToAnswer := map[string]string{}
	for _, res := range report.Content.Results {
		questionToAnswer[res.Question] = res.Answer
	}

	if len(questionToAnswer) != len(expectedAnswers) {
		t.Fatal("incorrect number of questions answered")
	}

	for question, expected := range expectedAnswers {
		answer, ok := questionToAnswer[question]
		if !ok || !strings.Contains(answer, expected) {
			t.Fatalf("incorrect answer '%v' for question '%v'", answer, question)
		}
	}

	if reuseReportId != nil {
		err = ke.DeleteReport(reportId)
		if err != nil {
			t.Fatal(err)
		}

		_, err = ke.GetReport(reportId)
		if err == nil || !strings.Contains(err.Error(), "status 404") {
			t.Fatal("report should return 404")
		}
	}

	return reportId
}

func TestKnowledgeExtraction(t *testing.T) {
	c := getClient(t)

	ke, err := c.CreateKnowledgeExtractionWorkflow(randomName("knowledge-extraction"), []string{
		"net revenue of apple",
		"iphone sales in 2021 (in billion)",
		"a question that should be deleted",
		"did sales in europe change from 2022 to 2023",
		"how much did apple spend on research and development in 2021",
	}, "openai", true)
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &ke.ModelClient, false)

	err = ke.AddQuestion("what were the EPS in 2022")
	if err != nil {
		t.Fatal(err)
	}

	questions, err := ke.ListQuestions()
	if err != nil {
		t.Fatal(err)
	}

	if len(questions) != 6 {
		t.Fatalf("invalid question list: %v", questions)
	}

	for _, q := range questions {
		checkAndProcessQuestion(t, ke, q)
	}

	expectedAnswers := map[string]string{
		"net revenue of apple":                                         "383.3",
		"iphone sales in 2021 (in billion)":                            "191.973",
		"did sales in europe change from 2022 to 2023":                 "decreased",
		"how much did apple spend on research and development in 2021": "21,914",
		"what were the EPS in 2022":                                    "6.15",
	}

	reusedReportId := checkReports(t, ke, expectedAnswers, nil)

	badReportId, err := ke.CreateReport([]client.FileInfo{
		{Path: "./utils.go", Location: "upload"},
	})
	if err != nil {
		t.Fatal(err)
	}

	badReport, err := ke.AwaitReport(badReportId, 20*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	expectedMsg := "Error processing report: Unable to process document 'utils.go'. Please ensure that document is a supported type (pdf, docx, csv, html) and has correct extension."
	if badReport.Status != "failed" || badReport.Msg != expectedMsg {
		t.Fatal("invalid contents of failed report")
	}

	knowledgeExtractionModelIds := []uuid.UUID{ke.GetModelID()}

	apiKeyName := fmt.Sprintf("test-api-key-%s", ke.GetModelID().String())
	now := time.Now()

	expiry := now.Add(24 * time.Hour)

	keApiKey, err := c.CreateAPIKey(knowledgeExtractionModelIds, apiKeyName, expiry, false)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	if keApiKey == "" {
		t.Fatal("Expected a valid API key, but got an empty string")
	}

	err = ke.ModelClient.UseApiKey(keApiKey)
	if err != nil {
		t.Fatal("error using api_key in place of auth token.")
	}

	questions, err = ke.ListQuestions()
	if err != nil {
		t.Fatal(err)
	}

	if len(questions) != 5 {
		t.Fatalf("invalid question list: %v", questions)
	}

	for _, q := range questions {
		checkAndProcessQuestion(t, ke, q)
	}

	checkReports(t, ke, expectedAnswers, &reusedReportId)
}
