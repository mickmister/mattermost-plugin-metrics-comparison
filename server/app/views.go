package app

import (
	"fmt"
	"strings"
)

func (report *DBStoreReport) AsMarkdown() string {
	resText := "## Percentage Differences of DB Store Methods\n\n"

	resText += "Parameters:\n"
	resText += report.RunFlags.AsMarkdown()

	resText += "\n\n### Biggest Increases\n\n"
	resText += generateMarkdownTable(report.BiggestIncreases)

	resText += "\n\n### Biggest Decreases\n\n"
	resText += generateMarkdownTable(report.BiggestDecreases)

	return resText
}

func (runFlags RunReportFlags) AsMarkdown() string {
	bullets := []string{"- First time frame start: " + runFlags.First}
	bullets = append(bullets, "- Second time frame start: "+runFlags.Second)
	bullets = append(bullets, "- Time frame length: "+runFlags.Length)
	return strings.Join(bullets, "\n")
}

func generateMarkdownTable(entries []*DBEntry) string {
	resText := "| Method | Total Time | Count | Average |\n"
	resText += "| -------- | -------- | -------- | -------- |\n"
	for _, v := range entries {
		resText += fmt.Sprintf("|%s | %v | %v | %v |\n", v.Method, v.TotalTime, v.Count, v.Average)
	}

	return resText
}
