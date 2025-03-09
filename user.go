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
	Question ChatMessage `json:"question"`
	Answer   ChatMessage `json:"answer"`
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
	ctx map[string]any

	rwm sync.RWMutex
}

func newUser(id int64) *user {
	user := &user{
		id:    id,
		last:  time.Now(),
		model: deepseek.DeepSeekChat,
		ctx:   make(map[string]any),
	}
	err := user.initDir()
	if err != nil {
		log.Println("[warning] failed to initialize user data directory:", err)
	}
	user.readCharacter()
	user.readConversation()
	return user
}

func (user *user) initDir() error {
	for _, path := range []string{
		fmt.Sprintf("data/characters/%d", user.id),
		fmt.Sprintf("data/conversation/%d", user.id),
		fmt.Sprintf("data/memory/private/%d", user.id),
	} {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (user *user) readCharacter() {
	path := fmt.Sprintf("data/characters/%d/current.cfg", user.id)
	role, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if len(role) == 0 {
		return
	}
	path = fmt.Sprintf("data/characters/%d/%s.txt", user.id, role)
	char, err := os.ReadFile(path)
	if err != nil {
		log.Println("[error] failed to read character file:", err)
		return
	}
	user.role = string(role)
	user.character = string(char)
}

func (user *user) readConversation() {
	path := fmt.Sprintf("data/conversation/%d/current.json", user.id)
	stat, err := os.Stat(path)
	if err != nil {
		return
	}
	// skip old conversation
	if time.Since(stat.ModTime()) > conversationTimeout {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var rounds []*round
	err = jsonDecode(data, &rounds)
	if err != nil {
		log.Println("failed to decode current conversation:", err)
		return
	}
	user.rounds = rounds
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
	if time.Since(user.last) > conversationTimeout {
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

func (user *user) getContext(key string) any {
	user.rwm.RLock()
	defer user.rwm.RUnlock()
	return user.ctx[key]
}

func (user *user) setContext(key string, data any) {
	user.rwm.Lock()
	defer user.rwm.Unlock()
	user.ctx[key] = data
}
