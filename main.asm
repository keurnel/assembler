; ==============================================================================
; Keurnel Assembly Language - X86_64 Syntax Example
; ==============================================================================
; This file demonstrates the syntax of Keurnel Assembly Language (KASM)
; for documentation and syntax highlighting purposes.
; ==============================================================================

section .text
global _start

; ==============================================================================
; Main Entry Point
; ==============================================================================
_start:
    ; Initialize and call math functions
    mov rdi, 5                      ; First argument: a = 5
    mov rsi, 3                      ; Second argument: b = 3
    call math_add                   ; Call math::_add(5, 3)

    mov rdi, 10                     ; First argument: a = 10
    mov rsi, 4                      ; Second argument: b = 4
    call math_subtract              ; Call math::_subtract(10, 4)

    ; Call local example function
    call _start_example

    ; Call lifecycle initialization
    call lifecycle_start

    ; Exit program gracefully
    mov rax, 60                     ; syscall: exit
    xor rdi, rdi                    ; status: 0
    syscall

; ==============================================================================
; Local Functions (Global Namespace)
; ==============================================================================

_start_example:
    ; Example function in the global scope
    ; Demonstrates local function calls
    push rbp
    mov rbp, rsp

    ; Function body
    mov rax, 0x1234
    mov rbx, rax

    pop rbp
    ret

global_group:
    ; Initialize global state
    xor rax, rax                    ; Set rax to 0
    xor rbx, rbx                    ; Set rbx to 0
    xor rcx, rcx                    ; Set rcx to 0
    ret

; ==============================================================================
; Namespace: math
; Mathematical operations and utilities
; ==============================================================================

math_start:
    ; Initialize math subsystem
    mov rdi, 0                      ; Clear first argument
    mov rsi, 1                      ; Set second argument to 1
    call math_add
    ret

math_add:
    ; Add two 64-bit integers
    ; Parameters:
    ;   rdi = a (first operand)
    ;   rsi = b (second operand)
    ; Returns:
    ;   rax = result (a + b)
    mov rax, rdi                    ; Copy first operand to rax
    add rax, rsi                    ; Add second operand
    ret

math_subtract:
    ; Subtract two 64-bit integers
    ; Parameters:
    ;   rdi = a (minuend)
    ;   rsi = b (subtrahend)
    ; Returns:
    ;   rax = result (a - b)
    mov rax, rdi                    ; Copy minuend to rax
    sub rax, rsi                    ; Subtract subtrahend
    ret

math_multiply:
    ; Multiply two 64-bit integers
    ; Parameters:
    ;   rdi = a (multiplicand)
    ;   rsi = b (multiplier)
    ; Returns:
    ;   rax = result (a * b)
    mov rax, rdi                    ; Copy multiplicand to rax
    imul rax, rsi                   ; Multiply by multiplier
    ret

math_divide:
    ; Divide two 64-bit integers
    ; Parameters:
    ;   rdi = a (dividend)
    ;   rsi = b (divisor)
    ; Returns:
    ;   rax = quotient
    ;   rdx = remainder
    mov rax, rdi                    ; Copy dividend to rax
    xor rdx, rdx                    ; Clear rdx for division
    idiv rsi                        ; Divide by divisor
    ret

; ==============================================================================
; Namespace: lifecycle
; System lifecycle management functions
; ==============================================================================

lifecycle_start:
    ; Initialize lifecycle subsystem
    ; Uses math namespace functions for demonstration
    push rbp
    mov rbp, rsp

    ; Perform initialization calculation
    mov rdi, 100                    ; Set a = 100
    mov rsi, 50                     ; Set b = 50
    call math_add                   ; Call math::_add(100, 50)

    ; Store result
    mov r12, rax                    ; Save result in r12

    ; Perform another operation
    mov rdi, r12                    ; Use previous result
    mov rsi, 25                     ; Set b = 25
    call math_subtract              ; Call math::_subtract(result, 25)

    pop rbp
    ret

lifecycle_shutdown:
    ; Cleanup and shutdown
    ; Clear all general-purpose registers
    xor rax, rax
    xor rbx, rbx
    xor rcx, rcx
    xor rdx, rdx
    xor rsi, rsi
    xor rdi, rdi
    xor r8, r8
    xor r9, r9
    xor r10, r10
    xor r11, r11
    ret

lifecycle_restart:
    ; Restart the lifecycle
    call lifecycle_shutdown
    call lifecycle_start
    ret

; ==============================================================================
; Advanced Features Demonstration
; ==============================================================================

demo_registers:
    ; Demonstrate various register operations
    ; 64-bit registers
    mov rax, 0x1234567890ABCDEF
    mov rbx, rax

    ; 32-bit registers
    mov eax, 0x12345678
    mov ebx, eax

    ; 16-bit registers
    mov ax, 0x1234
    mov bx, ax

    ; 8-bit registers
    mov al, 0x12
    mov bl, al

    ret

demo_memory:
    ; Demonstrate memory operations
    push rbp
    mov rbp, rsp
    sub rsp, 16                     ; Allocate 16 bytes on stack

    ; Store values
    mov qword [rbp-8], 0x1234
    mov qword [rbp-16], 0x5678

    ; Load values
    mov rax, qword [rbp-8]
    mov rbx, qword [rbp-16]

    ; Add and store
    add rax, rbx
    mov qword [rbp-8], rax

    add rsp, 16                     ; Deallocate stack space
    pop rbp
    ret

demo_conditionals:
    ; Demonstrate conditional operations
    mov rax, 10
    mov rbx, 20

    cmp rax, rbx                    ; Compare rax and rbx
    jl .less_than                   ; Jump if less
    jmp .end                        ; Otherwise jump to end

.less_than:
    mov rcx, 1                      ; Set flag

.end:
    ret

demo_loops:
    ; Demonstrate loop operations
    mov rcx, 10                     ; Counter = 10
    xor rax, rax                    ; Sum = 0

.loop:
    add rax, rcx                    ; Add counter to sum
    dec rcx                         ; Decrement counter
    jnz .loop                       ; Jump if not zero

    ret

; ==============================================================================
; Data Section (for reference)
; ==============================================================================

section .data
    msg: db "Hello, Keurnel Assembly!", 0x0A
    msg_len: equ $ - msg

    number: dq 0x1234567890ABCDEF
    array: dq 1, 2, 3, 4, 5, 6, 7, 8

section .bss
    buffer: resb 256
    result: resq 1

; ==============================================================================
; End of Keurnel Assembly Example
; ==============================================================================

