; =============================================================================
; x86_64 Bootloader - Hello World
; =============================================================================
; Stage 1: 16-bit real mode bootloader (512 bytes, sector 1)
;   - Initializes segments and stack
;   - Loads stage 2 (het programma) van disk naar geheugen op 0x7E00
;   - Springt naar stage 2
;
; Stage 2: Het programma dat "Hello, World!" weergeeft (sector 2)
; =============================================================================

[BITS 16]
[ORG 0x7C00]

; ---------------------
; Stage 1: Bootloader
; ---------------------
stage1_start:
    cli                     ; Interrupts uit tijdens setup
    xor ax, ax
    mov ds, ax
    mov es, ax
    mov ss, ax
    mov sp, 0x7C00          ; Stack groeit naar beneden vanaf 0x7C00
    sti                     ; Interrupts weer aan

    ; Sla boot drive nummer op (BIOS geeft dit in DL)
    mov [boot_drive], dl

    ; Laad stage 2 (het programma) van disk naar geheugen
    ; Stage 2 staat in sector 2, wordt geladen op 0x7E00
    mov ah, 0x02            ; BIOS functie: lees sectoren
    mov al, 1               ; Aantal sectoren om te lezen
    mov ch, 0               ; Cylinder 0
    mov cl, 2               ; Sector 2 (1-indexed, sector 1 = bootloader)
    mov dh, 0               ; Head 0
    mov dl, [boot_drive]    ; Drive nummer
    mov bx, 0x7E00          ; Doel adres in geheugen (ES:BX)
    int 0x13                ; BIOS disk interrupt
    jc disk_error           ; Bij fout, spring naar error handler

    ; Spring naar het geladen programma
    jmp 0x0000:0x7E00

disk_error:
    mov si, err_msg
    call print_string_16
    jmp halt

print_string_16:
    lodsb                   ; Laad byte van [SI] in AL, verhoog SI
    or al, al               ; Is het null-terminator?
    jz .done
    mov ah, 0x0E            ; BIOS teletype functie
    mov bh, 0x00            ; Pagina 0
    int 0x10                ; Print karakter
    jmp print_string_16
.done:
    ret

halt:
    cli
    hlt
    jmp halt

; Data
boot_drive: db 0
err_msg:    db "Disk leesfout!", 0x0D, 0x0A, 0

; Vul aan tot 510 bytes en voeg boot signature toe
times 510 - ($ - $$) db 0
dw 0xAA55                  ; Boot signature

; =============================================================================
; Stage 2: Het programma - "Hello, World!"
; Wordt geladen op adres 0x7E00
; =============================================================================
[BITS 16]

program_start:
    ; Wis het scherm
    mov ah, 0x00            ; Video functie: stel video modus in
    mov al, 0x03            ; 80x25 tekst modus, 16 kleuren
    int 0x10

    ; Stel cursor positie in (rij 0, kolom 0)
    mov ah, 0x02
    mov bh, 0x00            ; Pagina 0
    mov dh, 0x00            ; Rij 0
    mov dl, 0x00            ; Kolom 0
    int 0x10

    ; Print de welkomstboodschap
    mov si, welcome_msg
    call print_color

    ; Print "Hello, World!" in felle kleuren
    mov si, hello_msg
    call print_color

    ; Print extra info
    mov si, info_msg
    call print_string_stage2

    ; Wacht op toetsaanslag
    mov si, press_key_msg
    call print_string_stage2
    xor ah, ah
    int 0x16                ; Wacht op toets

    ; Herstart na toetsaanslag
    int 0x19

print_color:
    ; Print string met kleur attribuut
    ; SI = pointer naar string
.loop:
    lodsb
    or al, al
    jz .done
    mov ah, 0x0E
    mov bl, 0x0A            ; Lichtgroen kleur
    mov bh, 0x00
    int 0x10
    jmp .loop
.done:
    ret

print_string_stage2:
    lodsb
    or al, al
    jz .done
    mov ah, 0x0E
    mov bh, 0x00
    int 0x10
    jmp print_string_stage2
.done:
    ret

; Data voor stage 2
welcome_msg:    db "=== Keurnel Bootloader ===", 0x0D, 0x0A, 0
hello_msg:      db "Hello, World!", 0x0D, 0x0A, 0
info_msg:       db 0x0D, 0x0A
                db "Bootloader succesvol geladen!", 0x0D, 0x0A
                db "Programma draait vanuit geheugen op 0x7E00", 0x0D, 0x0A, 0
press_key_msg:  db 0x0D, 0x0A, "Druk op een toets om te herstarten...", 0x0D, 0x0A, 0

; Vul stage 2 aan tot 512 bytes
times 1024 - ($ - $$) db 0

