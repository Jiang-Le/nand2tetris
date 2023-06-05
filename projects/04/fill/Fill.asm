// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

// Put your code here.

// R0用于记录当前是否有键盘输入
@R0
M=0
// R1记录当前屏幕写入的位置
@16384
D=A
@R1
M=D
@R2
M=0

(BEGIN)
// 读取键盘输入，并判断是否为空
@R0
M=1
@24576
D=M
@HAVE_INPUT
D;JNE
@R0
M=0
(HAVE_INPUT)

// 在当前屏幕的位置写入值
@R0
D=M
@R1
A=M
D=!D
D=D+1
M=D

// 根据键盘输入，决定当前屏幕位置往前移，还是后退
@R1
M=M+1
@R0
D=M
@KEEP_FORWARD
D;JNE
@R1
M=M-1
M=M-1
(KEEP_FORWARD)

// 如果R1小于16384，则修改R1位16384
@R1
D=M
@16384
D=D-A
@BIGGER_16384
D;JGE
@16384
D=A
@R1
M=D
(BIGGER_16384)

// 如果R1大于24608，则修改R1位16384
@R1
D=M
@24608
D=D-A
@SMALLER_24608
D;JLT
@16384
D=A
@R1
M=D
(SMALLER_24608)

@BEGIN
0;JMP