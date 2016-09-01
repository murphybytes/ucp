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

	packet, e := buildPacket(testData)
	assert.Nil(t, e)
	assert.Equal(t, len(testData)+headerSize, len(packet))

	e = compareChecksum(packet)
	assert.Nil(t, e)
	busted := make([]byte, len(packet))
	copy(busted, packet)
	busted[500] = busted[500] + 3
	e = compareChecksum(busted)
	assert.NotNil(t, e)

	busted = make([]byte, len(packet)-5)
	copy(busted, packet)
	e = compareChecksum(busted)
	assert.NotNil(t, e)

	size, _ := getBufferSize(packet)
	assert.Equal(t, 1000, size)

	data, _ := getDataPacket(packet)
	comp := bytes.Compare(data, testData)
	assert.Equal(t, 0, comp)

}
