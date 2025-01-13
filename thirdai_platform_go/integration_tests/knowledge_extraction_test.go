package integrationtests

import (
	"strings"
	"testing"
	"thirdai_platform/model_bazaar/config"
	"time"
)

func TestKnowledgeExtraction(t *testing.T) {
	c := getClient(t)

	client, err := c.CreateKnowledgeExtractionWorkflow(randomName("knowledge-extraction"), []string{
		"net revenue of apple",
		"iphone sales in 2021 (in billion)",
		"a question that should be deleted",
		"did sales in europe change from 2022 to 2023",
		"how much did apple spend on research and development in 2021",
	}, "openai")
	if err != nil {
		t.Fatal(err)
	}

	deployModel(t, &client.ModelClient, false)

	err = client.AddQuestion("what were the EPS in 2022")
	if err != nil {
		t.Fatal(err)
	}

	questions, err := client.ListQuestions()
	if err != nil {
		t.Fatal(err)
	}

	if len(questions) != 6 {
		t.Fatalf("invalid question list: %v", questions)
	}

	for _, q := range questions {
		if strings.Contains(q.QuestionText, "EPS") {
			err := client.AddKeywords(q.QuestionId, []string{"earnings", "per", "share"})
			if err != nil {
				t.Fatal(err)
			}
		}
		if strings.Contains(q.QuestionText, "deleted") {
			err := client.DeleteQuestion(q.QuestionId)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	reportId, err := client.CreateReport([]config.FileInfo{
		{Path: "./data/apple-10k.pdf",
			Location: "local",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	reports, err := client.ListReports()
	if err != nil {
		t.Fatal(err)
	}

	if len(reports) != 1 || reports[0].ReportId != reportId {
		t.Fatalf("incorrect reports returned: %v", reports)
	}

	report, err := client.AwaitReport(reportId, 200*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	expectedAnswers := map[string]string{
		"net revenue of apple":                                         "383.3",
		"iphone sales in 2021 (in billion)":                            "191.973",
		"did sales in europe change from 2022 to 2023":                 "decreased",
		"how much did apple spend on research and development in 2021": "21,914",
		"what were the EPS in 2022":                                    "6.15",
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

	err = client.DeleteReport(reportId)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetReport(reportId)
	if err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Fatal("report should return 404")
	}

	badReportId, err := client.CreateReport([]config.FileInfo{
		{Path: "./utils.go", Location: "local"},
	})

	badReport, err := client.AwaitReport(badReportId, 20*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	expectedMsg := "Error processing report: Unable to process document 'utils.go'. Please ensure that document is a supported type (pdf, docx, csv, html) and has correct extension."
	if badReport.Status != "failed" || badReport.Msg != expectedMsg {
		t.Fatal("invalid contents of failed report")
	}
}
