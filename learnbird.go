package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Set struct {
	words    []Word
	index    int
	len      int
	mistakes int
}

var state = make(map[int64]Set)
var mutex = &sync.Mutex{}

func randArray(words []Word) {
	rand.Seed(time.Now().UnixNano())
	// Fisher-Yates algorithm
	for i := len(words) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		words[i], words[j] = words[j], words[i]
	}
}

func increaseState(idUser int64, client *mongo.Client, depth int64) {
	words := getData(client, "user_"+strconv.FormatInt(idUser, 10), depth)
	randArray(words)
	mutex.Lock()
	defer mutex.Unlock()
	state[idUser] = Set{words, 0, len(words), 0}
}

func decreaseState(idUser int64) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(state, idUser)
}

func askWord(idUser int64) string {
	if state[idUser].index >= state[idUser].len {
		return "No task!"
	}
	out := "Please, translate this: \n"
	out = out + state[idUser].words[state[idUser].index].Rus
	return out
}

func prepareCompare(word string) string {
	listSymbols := []string{".", ",", "!", "?", " "}

	out := strings.ToLower(word)
	for _, symbol := range listSymbols {
		out = strings.Trim(out, symbol)
	}
	return out
}

func compareSentences(strOriginal string, strReference string) (string, bool) {
	out := "Wrong! Correct: \n"
	flag := true

	strOriginal = strings.ReplaceAll(strOriginal, "   ", " ")
	strOriginal = strings.ReplaceAll(strOriginal, "  ", " ")
	arrOriginal := strings.Split(strOriginal, " ")
	arrReference := strings.Split(strReference, " ")
	if len(arrOriginal) != len(arrReference) {
		return fmt.Sprintf("%s <b>%s</b>", out, strReference), false
	}
	for i, word := range arrOriginal {
		if prepareCompare(word) != prepareCompare(arrReference[i]) {
			out = out + fmt.Sprintf("<b>%s</b> ", arrReference[i])
			flag = false
		} else {
			out = out + fmt.Sprintf("%s ", arrReference[i])
		}
	}
	return out, flag
}

func checkAnswer(idUser int64, word string) string {
	mistake := 0
	var out string
	if set, ok := state[idUser]; ok {
		if out, ok = compareSentences(word, set.words[set.index].Eng); ok {
			out = "Wow, correct answer! \n"
		} else {
			mistake = 1
		}
		set.index++
		set.mistakes += mistake
		mutex.Lock()
		state[idUser] = set
		mutex.Unlock()
	} else {
		return "something wrong, try to write /help"
	}
	if state[idUser].index >= state[idUser].len {
		out = fmt.Sprintf("%s \nYou have finished the task! \n %d mistakes from %d", out, state[idUser].mistakes, state[idUser].len)
		decreaseState(idUser)
	} else {
		out = fmt.Sprintf("%s \n %s", out, askWord(idUser))
	}
	return out
}
