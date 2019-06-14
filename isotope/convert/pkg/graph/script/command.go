// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package script

import (
	"encoding/json"
	"fmt"
)

// Command is the top level interface for commands.
type Command interface{}

const (
	sleepCommandKey   = "sleep"
	requestCommandKey = "call"
	processCommandKey = "process"
)

func commandsToMarshallable(cmds []Command) ([]interface{}, error) {
	marshallableCmds := make([]interface{}, 0, len(cmds))
	for _, cmd := range cmds {
		marshallableCmd, err := commandToMarshallable(cmd)
		if err != nil {
			return nil, err
		}
		marshallableCmds = append(marshallableCmds, marshallableCmd)
	}
	return marshallableCmds, nil
}

func commandToMarshallable(cmd Command) (interface{}, error) {
	switch cmd := cmd.(type) {
	case SleepCommand:
		return map[string]string{sleepCommandKey: cmd.String()}, nil
	case RequestCommand:
		return map[string]RequestCommand{requestCommandKey: cmd}, nil
	case ConcurrentCommand:
		return commandsToMarshallable(cmd)
	case ProcessCommand:
		return map[string]ProcessCommand{processCommandKey: cmd}, nil
	default:
		return nil, InvalidCommandTypeError{cmd}
	}
}

func parseJSONCommands(b []byte) ([]Command, error) {
	var wrappedCmds []unmarshallableCommand
	err := json.Unmarshal(b, &wrappedCmds)
	if err != nil {
		return nil, err
	}

	cmds := make([]Command, 0, len(wrappedCmds))
	for _, wrappedCmd := range wrappedCmds {
		cmd := wrappedCmd.Command
		cmds = append(cmds, cmd)
	}
	return cmds, nil
}

// unmarshallableCommand wraps a Command so that it may act as a receiver.
type unmarshallableCommand struct{ Command }

func (c *unmarshallableCommand) UnmarshalJSON(b []byte) error {
	// This function is called after ghodss/yaml converts YAML to _sanitized_
	// JSON, so we assume the JSON is valid and does not have leading spaces.
	isJSONArray := b[0] == '['
	if isJSONArray {
		var concurrentCommand ConcurrentCommand
		err := json.Unmarshal(b, &concurrentCommand)
		if err != nil {
			return err
		}
		c.Command = concurrentCommand
	} else {
		key, err := parseJSONCommandKey(b)
		if err != nil {
			return err
		}
		switch key {
		case sleepCommandKey:
			c.Command, err = parseSleepCommandFromJSONMap(b)
			if err != nil {
				return err
			}
		case requestCommandKey:
			c.Command, err = parseRequestCommandFromJSONMap(b)
			if err != nil {
				return err
			}
		case processCommandKey:
			c.Command, err = parseProcessCommandFromJSONMap(b)
			if err != nil {
				return err
			}
		default:
			return UnknownCommandKeyError{key}
		}
	}
	return nil
}

func parseJSONCommandKey(b []byte) (s string, err error) {
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	if len(m) > 1 {
		err = MultipleKeysInCommandMapError{m}
		return
	}
	// Should only loop once, setting s to the single command key in the map.
	for s = range m {
	}
	return
}

// b must contain a single key whose value is an unmarshallable ProcessCommand.
func parseProcessCommandFromJSONMap(b []byte) (cmd ProcessCommand, err error) {
	var m map[string]ProcessCommand
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	for _, cmd = range m {
	}
	return
}

// b must contain a single key whose value is an unmarshallable SleepCommand.
func parseSleepCommandFromJSONMap(b []byte) (cmd SleepCommand, err error) {
	var m map[string]SleepCommand
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	for _, cmd = range m {
	}
	return
}

// b must contain a single key whose value is an unmarshallable RequestCommand.
func parseRequestCommandFromJSONMap(b []byte) (cmd RequestCommand, err error) {
	var m map[string]RequestCommand
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	for _, cmd = range m {
	}
	return
}

// InvalidCommandTypeError is returned when a type-switch on a Command does not
// reveal a known Command.
type InvalidCommandTypeError struct {
	Command Command
}

func (e InvalidCommandTypeError) Error() string {
	return fmt.Sprintf("invalid command type: %T", e.Command)
}

// MultipleKeysInCommandMapError is returned when there is more than one key in
// a command map.
type MultipleKeysInCommandMapError struct {
	CommandMap map[string]interface{}
}

func (e MultipleKeysInCommandMapError) Error() string {
	return fmt.Sprintf("multiple keys for command: %v", e.CommandMap)
}

// UnknownCommandKeyError is returned when a command's key (i.e. "sleep") does
// not match a known command.
type UnknownCommandKeyError struct {
	CommandKey string
}

func (e UnknownCommandKeyError) Error() string {
	return fmt.Sprintf("unknown command: %s", e.CommandKey)
}
