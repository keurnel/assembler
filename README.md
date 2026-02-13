# Keurnel - Assembler
Keurnels assembler is a tool that converts assembly language code into machine code that can be executed by keurnel.


## Keurnel Assembler Syntax
The syntax of keurnel assembler is similar to other assemblers, with some specific instructions and directives that are
unique to keurnel.

### Main Directive
Each program has a entry-point which is defined using the `.main:` directive.

**Example:**
```asm
.main:
    ; Your code here
    ret
```