; Global start directive.
.start:
    MOV AX, 0x0000 ; Load the value 0x0000 into the AX register.
    ret

namespace my_namespace

; Namespace specific start directive.
.start:
    ; Entry-point of the program.
    MOV AX, 0x1234 ; Load the value 0x1234 into the AX register.
    MOV BX, 0x5678 ; Load the value 0x5678 into the BX register.
    ADD AX, BX      ; Add the value in BX to AX, result
    ret

; This is another group.
group:
    MOV CX, 0x9ABC ; Load the value 0x9ABC into the CX register.
    SUB CX, AX      ; Subtract the value in AX from CX, result
    ret

; This is a third group.
.another_group:
    MOV DX, 0xDEF0 ; Load the value 0xDEF0 into the DX register.
    ret


MOV AX, 0x1111 ; Load the value 0x1111 into the AX register.

namespace another_namespace

; Another namespace specific start directive.
.start:
    MOV AX, 0x2222 ; Load the value 0x2222 into the AX register.
    ret

MOV AX, 0x3333 ; Load the value 0x3333 into the AX register.