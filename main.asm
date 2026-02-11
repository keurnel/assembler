mov rax, 42      ; Move the value 42 into rax
mov rbx, rax     ; Move the value in rax (42) into rbx
; Exit the program
mov rax, 60      ; syscall: exit
xor rdi, rdi     ; status: 0
syscall            ; call kernel