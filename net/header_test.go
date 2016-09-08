package net

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacketValidation(t *testing.T) {
	testData := make([]byte, 1000)
	rand.Read(testData)

	packet, e := prependHeaderToBuffer(testData)
	assert.Nil(t, e)
	assert.Equal(t, len(testData)+headerLen, len(packet))

	size, _ := getBufferSize(packet)
	assert.Equal(t, 1000, size)

	data, _ := getDataPacket(packet)
	comp := bytes.Compare(data, testData)
	assert.Equal(t, 0, comp)

}
