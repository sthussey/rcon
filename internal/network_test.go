package internal

import (
	"testing"
)

func TestGenerateNetNSName(t *testing.T) {
    t.Parallel()

    name, err := generateNetNSName()

    if err != nil {
        t.Errorf("Error generating netns name: %v", err)
    }

    var name2 string

    name2, err = generateNetNSName()

    if err != nil {
        t.Errorf("Error generating second netns name: %v", err)
    }

    if name == name2 {
        t.Errorf("Generated duplicated netns names: %s, %s", name, name2)
    }
}

