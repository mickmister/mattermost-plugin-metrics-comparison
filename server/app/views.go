package app

import (
	"fmt"
	"math"
	"strings"
)

func (report *DBStoreReport) AsMarkdown() string {
	resText := "## Percentage Differences of DB Store Methods\n\n"

	resText += "Parameters:\n"
	resText += report.RunFlags.AsMarkdown()

	resText += "\n\n### Biggest Increases\n\n"
	resText += generateDBMarkdownTable(report.BiggestIncreases)

	resText += "\n\n### Biggest Decreases\n\n"
	resText += generateDBMarkdownTable(report.BiggestDecreases)

	return resText
}

func generateDBMarkdownTable(entries []*FullDBEntry) string {
	resText := "| Method | Total Diff | Total 1 | Total 2 | Count Diff | Count 1 | Count 2 | Average Diff | Average 1 | Average 2 |\n"
	resText += "| -------- | -------- | -------- | -------- |\n"
	for _, v := range entries {
		resText += fmt.Sprintf("|%s | %v%% | %.3f | %.3f | %v%% | %v | %v | %v%% | %.5f | %.5f |\n",
			v.Method,
			v.TotalTimeDifferencePercent,
			v.Timeframe1.TotalTime,
			v.Timeframe2.TotalTime,
			v.CountDifferencePercent,
			math.Ceil(v.Timeframe1.Count),
			math.Ceil(v.Timeframe2.Count),
			v.AverageDifferencePercent,
			v.Timeframe1.Average,
			v.Timeframe2.Average,
		)
	}

	return resText
}

func (report *APIHandlerReport) AsMarkdown() string {
	resText := "## Percentage Differences of API Handlers\n\n"

	resText += "Parameters:\n"
	resText += report.RunFlags.AsMarkdown()

	resText += "\n\n### Biggest Increases\n\n"
	resText += generateAPIMarkdownTable(report.BiggestIncreases)

	resText += "\n\n### Biggest Decreases\n\n"
	resText += generateAPIMarkdownTable(report.BiggestDecreases)

	return resText
}

func generateAPIMarkdownTable(entries []*FullAPIEntry) string {
	resText := "| Handler | Total Diff | Total 1 | Total 2 | Count Diff | Count 1 | Count 2 | Average Diff | Average 1 | Average 2 |\n"
	resText += "| -------- | -------- | -------- | -------- |\n"
	for _, v := range entries {
		resText += fmt.Sprintf("|%s | %v%% | %.3f | %.3f | %v%% | %v | %v | %v%% | %.5f | %.5f |\n",
			v.Handler,
			v.TotalTimeDifferencePercent,
			v.Timeframe1.TotalTime,
			v.Timeframe2.TotalTime,
			v.CountDifferencePercent,
			math.Ceil(v.Timeframe1.Count),
			math.Ceil(v.Timeframe2.Count),
			v.AverageDifferencePercent,
			v.Timeframe1.Average,
			v.Timeframe2.Average,
		)
	}

	return resText
}

func (runFlags RunReportFlags) AsMarkdown() string {
	bullets := []string{"- First time frame start: " + runFlags.First}
	bullets = append(bullets, "- Second time frame start: "+runFlags.Second)
	bullets = append(bullets, "- Time frame length: "+runFlags.Length)
	return strings.Join(bullets, "\n")
}
