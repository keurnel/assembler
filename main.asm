; This is a simple comment.
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