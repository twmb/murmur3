// +build go1.5,amd64

// Sum32(data []byte) (h1 uint32)
TEXT ·Sum32(SB), $0-28
	MOVQ data_base+0(FP), SI
	MOVQ data_len+8(FP), R9
	XORL R12, R12
	LEAQ h1+24(FP), BX
	JMP  sum32internal<>(SB)

// SeedSum32(seed uint32, data []byte) (h1 uint32)
TEXT ·SeedSum32(SB), $0-36
	MOVL seed+0(FP), R12
	MOVQ data_base+8(FP), SI
	MOVQ data_len+16(FP), R9
	LEAQ h1+32(FP), BX
	JMP  sum32internal<>(SB)

#define c1_32 0xcc9e2d51
#define c2_32 0x1b873593

// Expects:
// R12 == uint32 seed
// SI  == &data
// R9  == len(data)
// BX  == &uint32 return
TEXT sum32internal<>(SB), $0
	XORQ BP, BP
	MOVQ R9, CX
	ANDQ $-4, CX // cx = len(data) - (len(data) % 4)

	// for bp = 0 ; bp < cx; bp += 4 {...
loop:
	CMPQ BP, CX
	JE   tail
	MOVL (SI)(BP*1), AX
	ADDQ $4, BP

	IMULL $c1_32, AX
	ROLL  $15, AX
	IMULL $c2_32, AX

	XORL AX, R12
	ROLL $13, R12
	LEAL 0xe6546b64(R12)(R12*4), R12

	JMP loop

tail:
	MOVQ R9, CX
	ANDL $3, CX
	JZ   finalize

	XORQ AX, AX

	SUBQ $2, CX
	JL   tail1
	JZ   tail2

tail3:  // no jump
	MOVB 2(SI)(BP*1), AX
	SALL $16, AX

tail2:
	MOVW (SI)(BP*1), AX
	JMP  fintail

tail1:
	MOVB (SI)(BP*1), AL

fintail:
	IMULL $c1_32, AX
	ROLL  $15, AX
	IMULL $c2_32, AX
	XORL  AX, R12

finalize:
	XORL R9, R12

	// fmix32
	MOVL  R12, DX
	SHRL  $16, DX
	XORL  DX, R12
	IMULL $0x85ebca6b, R12
	MOVL  R12, DX
	SHRL  $13, DX
	XORL  DX, R12
	IMULL $0xc2b2ae35, R12
	MOVL  R12, DX
	SHRL  $16, DX
	XORL  DX, R12

	MOVL R12, (BX)
	RET
