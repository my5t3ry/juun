package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"math"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	. "../common"
	. "../config"

	. "github.com/jackdoe/juun/vw"
	iq "github.com/rekki/go-query"
	analyzer "github.com/rekki/go-query-analyze"
	index "github.com/rekki/go-query-index"
	log "github.com/sirupsen/logrus"
)

func toDocuments(in []*HistoryLine) []index.Document {
	out := make([]index.Document, len(in))
	for i, d := range in {
		out[i] = index.Document(d)
	}
	return out
}

type History struct {
	Lines        map[string]*HistoryLine
	SortedLines  []*HistoryLine
	idx          *index.MemOnlyIndex
	lock         sync.Mutex
	vw           *Bandit
	globalCursor int
}

func NewHistory() *History {

	m := index.NewMemOnlyIndex(map[string]*analyzer.Analyzer{
		"name":         index.AutocompleteAnalyzer,
		"name_fuzzy":   index.FuzzyAnalyzer,
		"name_soundex": index.SoundexAnalyzer,
		"country":      index.IDAnalyzer,
	})

	return &History{
		Lines: make(map[string]*HistoryLine), // ordered list of commands
		idx:   m,
	}
}

func (h *History) SelfReindex() {
	log.Infof("starting reindexing")

	h.SortedLines = h.flatMapLinesSorted()

	log.Infof("reindexing done, %d items", len(h.Lines))

}

func (h *History) filterLine(f func(string) bool) []*HistoryLine {

	lines := make([]*HistoryLine, 0, len(h.Lines))

	for _, value := range h.Lines {
		lines = append(lines, value)
	}

	filtered := make([]*HistoryLine, 0)
	for _, v := range lines {
		if f(v.Line) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func (h *History) flatMapLinesSorted() []*HistoryLine {

	lines := make([]*HistoryLine, 0, len(h.Lines))

	for _, value := range h.Lines {
		lines = append(lines, value)
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].TimeStamp > lines[j].TimeStamp
	})

	return lines
}

func (h *History) add(line string, env map[string]string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	now := time.Now().UnixNano()

	filtered := h.filterLine(func(word string) bool {
		return strings.Contains(word, line)
	})
	v := &HistoryLine{}
	if len(filtered) > 0 {
		v = filtered[0]
		v.Count++
		v.TimeStamp = now
		h.idx.Index(index.Document(v))
	} else {
		v = &HistoryLine{
			Line:      line,
			TimeStamp: now,
			Count:     1,
			Id:        len(h.Lines),
			Uuid:      uuid.NewString(),
		}

		h.Lines[v.Uuid] = v
		h.idx.Index(index.Document(v))
	}

	if h.vw != nil {
		h.vw.Click(v.Id)
		h.like(h.Lines[v.Uuid], env)
	}
	h.SelfReindex()
}

func (h *History) gotoend() {
	h.lock.Lock()
	defer h.lock.Unlock()

}
func (h *History) up(buf string) string {
	return h.move(true, buf)
}

func (h *History) down(buf string) string {
	return h.move(false, buf)
}

func (h *History) getTerminal() *Terminal {
	return nil
}
func (h *History) move(goUP bool, buf string) string {
	h.lock.Lock()
	defer h.lock.Unlock()

	if !goUP && h.globalCursor > 0 {
		h.globalCursor -= 1
	} else if goUP && h.globalCursor < len(h.Lines) {
		h.globalCursor += 1
	}

	if len(h.Lines) == 0 || !goUP && h.globalCursor == 0 {
		return ""
	}

	return h.SortedLines[h.globalCursor].Line
}
func (h *History) getLastLines() []*HistoryLine {
	h.lock.Lock()
	defer h.lock.Unlock()

	if len(h.Lines) == 0 {
		return nil
	}

	cfg := GetConfig()

	return h.SortedLines[:cfg.SearchResults]
}

func reverseLines(input []*HistoryLine) []*HistoryLine {
	if len(input) == 0 {
		return input
	}
	return append(reverseLines(input[1:]), input[0])
}

type scored struct {
	uuid          string
	id            int
	score         float32
	tfidf         float32
	countScore    float32
	timeScore     float32
	terminalScore float32
}

type ByScore []scored

func (s ByScore) Len() int {
	return len(s)
}
func (s ByScore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByScore) Less(i, j int) bool {
	return s[j].score < s[i].score
}

const scoreOnTerminal = float32(10)

func (h *History) search(text string, env map[string]string) []*HistoryLine {
	h.lock.Lock()
	defer h.lock.Unlock()

	text = strings.Trim(text, " ")
	if len(text) == 0 {
		return []*HistoryLine{}
	}
	query := iq.And(h.idx.Terms("line", text)...)

	score := []scored{}
	now := time.Now().Unix()

	h.idx.Foreach(query, func(did int32, tfidf float32, doc index.Document) {
		line := doc.(*HistoryLine)
		ts := line.TimeStamp / 1000000000
		timeScore := float32(-math.Sqrt(1 + float64(now-ts))) // -log(1+secondsDiff)

		countScore := float32(math.Sqrt(float64(line.Count)))
		terminalScore := float32(0)

		total := (3 * tfidf) + (timeScore) + 0 + countScore
		score = append(score, scored{uuid: line.Uuid, id: line.Id, score: total, tfidf: tfidf, timeScore: timeScore, terminalScore: terminalScore, countScore: countScore})

	})

	sort.Sort(ByScore(score))

	if h.vw != nil {
		// take the top 5 and sort them using vowpal wabbit's bootstrap
		topN := 2
		if topN > len(score) {
			topN = len(score)
		}

		ctx := UserContext(text, GetOrDefault(env, "cwd", ""))
		vwi := []*Item{}
		for i := 0; i < topN; i++ {
			s := score[i]
			line := h.Lines[s.uuid]

			f := line.Featurize()
			f.Add(ctx)
			f.AddNamespaces(
				NewNamespace("i_score",
					NewFeature("tfidf", s.tfidf),
					NewFeature("timeScore", s.timeScore),
					NewFeature("countScore", s.countScore),
					NewFeature(fmt.Sprintf("terminalScore=%d", int(s.terminalScore)), 0)))

			vwi = append(vwi, NewItem(line.Id, f.ToVW()))
			log.Debugf("before VW: tfidf: %f timeScore: %f terminalScore:%f countScore:%f line:%s", s.tfidf, s.timeScore, s.terminalScore, s.countScore, line.Line)
		}

		prediction := h.vw.Predict(1, vwi...)
		sort.Slice(score, func(i, j int) bool { return prediction[int(score[j].id)] < prediction[int(score[i].id)] })
	}

	// pick the first one
	out := []*HistoryLine{}
	if len(score) > 0 {
		for _, s := range score {
			line := h.Lines[s.uuid]
			out = append(out, line)
		}
	}
	cfg := GetConfig()
	if len(out) > cfg.SearchResults {
		out = out[:cfg.SearchResults]
	}
	return out
}

func (h *History) like(line *HistoryLine, env map[string]string) {
	if h.vw == nil {
		return
	}

	ctx := UserContext("", GetOrDefault(env, "cwd", ""))
	f := line.Featurize()
	f.Add(ctx)
	f.AddNamespaces(NewNamespace("i_score", NewFeature(fmt.Sprintf("terminalScore=%d", int(scoreOnTerminal)), 0)))
	h.vw.SendReceive(fmt.Sprintf("1 10 %s", f.ToVW())) // add weight of 10 on the clicked one
}

func (h *History) Save() {
	histfile := path.Join(GetHome(), ".juun.json")

	h.lock.Lock()
	d1, err := json.Marshal(h)
	h.lock.Unlock()
	if err == nil {
		SafeSave(histfile, func(tmp string) error {
			return ioutil.WriteFile(tmp, d1, 0600)
		})
	} else {
		log.Warnf("error marshalling: %s", err.Error())
	}

	if h.vw != nil {
		h.vw.Save()
	}
}

func (h *History) Load() {
	histfile := path.Join(GetHome(), ".juun.json")
	log.Infof("---------------------")
	log.Infof("loading %s", histfile)

	dat, err := ioutil.ReadFile(histfile)
	if err == nil {
		err = json.Unmarshal(dat, h)
		if err != nil {
			log.Warnf("err: %s", err.Error())
			h = NewHistory()
		}
	} else {
		log.Warnf("err: %s", err.Error())
	}

	h.SelfReindex()

	h.idx.Index(toDocuments(h.SortedLines)...)
}
