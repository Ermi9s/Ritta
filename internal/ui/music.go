package ui

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"

	"ritta/internal/logger"
)

//go:embed theme.mp3
var musicFile []byte

type MusicPlayer struct {
	cmd     *exec.Cmd
	file    string
	playing bool
	log     *logger.Logger
}

func NewMusicPlayer(log *logger.Logger) (*MusicPlayer, error) {
	file, err := os.CreateTemp("", "ritta-theme-*.mp3")
	if err != nil {
		log.Errorf("Failed to create temporary music file: %v", err)
		return nil, err
	}

	if _, err := file.Write(musicFile); err != nil {
		file.Close()
		os.Remove(file.Name())

		log.Errorf("Failed to write embedded music: %v", err)
		return nil, err
	}

	if err := file.Close(); err != nil {
		os.Remove(file.Name())

		log.Errorf("Failed to close music file: %v", err)
		return nil, err
	}

	log.Debug("Music player initialized")

	return &MusicPlayer{
		file: file.Name(),
		log:  log,
	}, nil
}

func (m *MusicPlayer) Play() error {
	if m.playing {
		return nil
	}

	var command string
	var args []string
	var playerName string

	if path, err := exec.LookPath("mpv"); err == nil {
		command = path
		playerName = "mpv"

		args = []string{
			"--no-video",
			"--really-quiet",
			m.file,
		}

	} else if path, err := exec.LookPath("pw-play"); err == nil {
		command = path
		playerName = "pw-play"

		args = []string{
			m.file,
		}

	} else {
		m.log.Warning("No audio player found; music disabled")
		return fmt.Errorf("no audio player found (tried mpv and pw-play)")
	}

	m.log.Debugf("Starting music player: %s", playerName)

	m.cmd = exec.Command(command, args...)

	if err := m.cmd.Start(); err != nil {
		m.log.Errorf("Failed to start music player: %v", err)
		return err
	}

	m.playing = true

	m.log.Info("Ritta theme music started")

	go func(cmd *exec.Cmd) {
		err := cmd.Wait()

		if m.cmd == cmd {
			m.playing = false
			m.cmd = nil
		}

		if err != nil {
			m.log.Debugf("Music player exited: %v", err)
		}
	}(m.cmd)

	return nil
}

func (m *MusicPlayer) Pause() error {
	if !m.playing || m.cmd == nil {
		return nil
	}

	m.log.Debug("Stopping Ritta theme music")

	if err := m.cmd.Process.Kill(); err != nil {
		m.log.Errorf("Failed to stop music: %v", err)
		return err
	}

	m.playing = false
	m.cmd = nil

	m.log.Info("Ritta theme music stopped")

	return nil
}

func (m *MusicPlayer) Toggle() error {
	if m.playing {
		return m.Pause()
	}

	return m.Play()
}

func (m *MusicPlayer) Playing() bool {
	return m.playing
}

func (m *MusicPlayer) Close() error {
	if m.cmd != nil {
		_ = m.cmd.Process.Kill()
		m.cmd = nil
	}

	m.playing = false

	if m.file != "" {
		if err := os.Remove(m.file); err != nil {
			m.log.Errorf("Failed to remove temporary music file: %v", err)
			return err
		}

		m.file = ""
	}

	m.log.Debug("Music player closed")

	return nil
}