package main

import (
	"fmt"
	"strconv"
	"sync"

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

func increaseState(idUser int64, client *mongo.Client, depth int64) {
	words := getData(client, "user_"+strconv.FormatInt(idUser, 10), depth)
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

func checkWord(idUser int64, word string) string {
	mistake := 0
	var out string
	if set, ok := state[idUser]; ok {
		if set.words[set.index].Eng != word {
			mistake = 1
			out = "Wrong! Correct: " + set.words[set.index].Eng
		} else {
			out = "Wow, correct answer! \n"
		}
		set.index++
		set.mistakes += mistake
		mutex.Lock()
		state[idUser] = set
		mutex.Unlock()
	} else {
		return "something wrong, try to write /help"
	}
	//test
	fmt.Println("index: ", state[idUser].index, " len: ", state[idUser].len)
	if state[idUser].index >= state[idUser].len {
		out = fmt.Sprintf("%s \nYou have finished the task! \n %d mistakes from %d", out, state[idUser].mistakes, state[idUser].len)
		//test
		fmt.Println(out)
		decreaseState(idUser)
		//test
		fmt.Println("after delete", out)
	} else {
		out = fmt.Sprintf("%s \n %s", out, askWord(idUser))
	}
	return out
}
