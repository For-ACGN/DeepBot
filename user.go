package deepbot

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cohesion-org/deepseek-go"
)

type round struct {
	Question ChatMessage
	Answer   ChatMessage
}

type user struct {
	id int64

	// current role name
	role string

	// current character content
	character string

	// chat context content
	rounds []*round
	last   time.Time

	// current model name
	model string

	// global disable tool call
	disableTC bool

	// about character mood
	mood string

	// store data for tool call
	ctx map[string]interface{}

	rwm sync.RWMutex
}

func newUser(id int64) *user {
	user := &user{
		id:    id,
		last:  time.Now(),
		model: deepseek.DeepSeekChat,
		ctx:   make(map[string]interface{}),
	}
	err := user.initDir()
	if err != nil {
		log.Println("[warning] failed to initialize user data directory:", err)
	}
	user.readCharacter()
	return user
}

func (user *user) initDir() error {
	err := os.MkdirAll(fmt.Sprintf("data/characters/%d", user.id), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fmt.Sprintf("data/memory/%d", user.id), 0755)
	if err != nil {
		return err
	}
	return nil
}

func (user *user) readCharacter() {
	role, err := os.ReadFile(fmt.Sprintf("data/characters/%d/current.cfg", user.id))
	if err != nil {
		return
	}
	if len(role) == 0 {
		return
	}
	char, err := os.ReadFile(fmt.Sprintf("data/characters/%d/%s.txt", user.id, role))
	if err != nil {
		log.Println("[error] failed to read character file:", err)
		return
	}
	user.role = string(role)
	user.character = string(char)
}

func (user *user) getRole() string {
	user.rwm.RLock()
	defer user.rwm.RUnlock()
	return user.role
}

func (user *user) getCharacter() string {
	user.rwm.RLock()
	defer user.rwm.RUnlock()
	return user.character
}

func (user *user) setCharacter(role, content string) {
	user.rwm.Lock()
	defer user.rwm.Unlock()
	user.role = role
	user.character = content
}

func (user *user) getRounds() []*round {
	user.rwm.Lock()
	defer user.rwm.Unlock()
	if time.Since(user.last) > 30*time.Minute {
		user.rounds = nil
	}
	user.last = time.Now()
	return user.rounds
}

func (user *user) setRounds(rounds []*round) {
	user.rwm.Lock()
	defer user.rwm.Unlock()
	user.rounds = rounds
	user.last = time.Now()
}

func (user *user) getModel() string {
	user.rwm.RLock()
	defer user.rwm.RUnlock()
	return user.model
}

func (user *user) setModel(model string) {
	user.rwm.Lock()
	defer user.rwm.Unlock()
	user.model = model
}

func (user *user) canToolCall() bool {
	user.rwm.RLock()
	defer user.rwm.RUnlock()
	return !user.disableTC
}

func (user *user) setToolCall(enabled bool) {
	user.rwm.Lock()
	defer user.rwm.Unlock()
	user.disableTC = !enabled
}

func (user *user) getMood() string {
	user.rwm.RLock()
	defer user.rwm.RUnlock()
	return user.mood
}

func (user *user) setMood(mood string) {
	user.rwm.Lock()
	defer user.rwm.Unlock()
	user.mood = mood
}

func (user *user) getContext(key string) interface{} {
	user.rwm.RLock()
	defer user.rwm.RUnlock()
	return user.ctx[key]
}

func (user *user) setContext(key string, data interface{}) {
	user.rwm.Lock()
	defer user.rwm.Unlock()
	user.ctx[key] = data
}
