package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	logNoTimestamp := log.New(os.Stderr, "", 0)
	bootRomFlag := flag.String("boot_rom", "", "the boot rom to use, optional")
	debugFlag := flag.Bool("debug", false, "start the emulator in debugger mode")
	traceFlag := flag.Bool("trace", false, "prints every executed instruction for debugging")
	muteFlag := flag.Bool("mute", false, "do not play sounds")
	flag.Parse()

	if flag.NArg() < 1 {
		logNoTimestamp.Fatal("A ROM file must be provided")
	}

	// Parse ROM and boot ROM (if provided)
	rom, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		logNoTimestamp.Fatal("Failed to load ROM: ", err)
	}
	var bootRom []byte
	if len(*bootRomFlag) > 0 {
		var err error
		bootRom, err = os.ReadFile(*bootRomFlag)
		if err != nil {
			logNoTimestamp.Fatal("Failed to load boot ROM: ", err)
		}
	}

	// Init game engine and emulator
	game := MakeGame()
	emulator := MakeEmulator(bootRom, rom, *debugFlag, *traceFlag, game)
	game.SetKeysListener(emulator.joypad)
	if !*muteFlag {
		game.SetAudioStream(emulator.apu)
	}

	// Start emulator and game. Emulator goes in a separate goroutine since it is blocking.
	go emulator.Run()
	game.Run()
}
