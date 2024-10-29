package main

const (
    addrNr10        = 0xff10
    addrNr11        = 0xff11
    addrNr12        = 0xff12
    addrNr13        = 0xff13
    addrNr14        = 0xff14
    addrNr21        = 0xff16
    addrNr22        = 0xff17
    addrNr23        = 0xff18
    addrNr24        = 0xff19
    addrNr30        = 0xff1a
    addrNr31        = 0xff1b
    addrNr32        = 0xff1c
    addrNr33        = 0xff1d
    addrNr34        = 0xff1e
    addrNr41        = 0xff20
    addrNr42        = 0xff21
    addrNr43        = 0xff22
    addrNr44        = 0xff23
    addrNr50        = 0xff24
    addrNr51        = 0xff25
    addrNr52        = 0xff26
    addrWavePattern = 0xff30
)

// Apu or Audio Processing Unit.
// This component is quite complicated with lots of quirks and edge cases, see:
// https://gbdev.gg8.se/wiki/articles/Gameboy_sound_hardware
type Apu struct {
}

func (a *Apu) OnTimer(divOld byte, divNew byte) {
    // TODO: implement
}

func (a *Apu) Get(addr uint16) (byte, bool) {
    // TODO: implement
    return 0, false
}

func (a *Apu) Set(addr uint16, v byte) bool {
    // TODO: implement
    return false
}