section .data
    msg1 db      "BEFORE breakpoint", 0xA
    len1 equ     $ - msg1
    msg2 db      "AFTER breakpoint", 0xA
    len2 equ     $ - msg2
    
section .text
    global _start
    
_start:
    ; Print BEFORE 
    mov     rax, 1
    mov     rdi, 1
    mov     rsi, msg1
    mov     rdx, len1
    syscall
    mov     rax, 60
    
    ; Print AFTER
    mov     rax, 1
    mov     rdi, 1 
    mov     rsi, msg2
    mov     rdx, len2
    syscall
    mov    rax, 60
    
    ; Setup exit code and return
    mov    rdi, 0
    syscall