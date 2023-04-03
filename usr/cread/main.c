#include <stdio.h>
#include <stdbool.h>
#include <unistd.h>
#include <string.h>

int main(){
    char buf[2];
    memset(buf, 0, 2);
    setvbuf(stdout, NULL, _IONBF, 0);
    setvbuf(stdin, NULL, _IONBF, 0);
    while(true){
        read(STDIN_FILENO, buf, 1);
        printf("%s", buf);
        if(buf[0] == '\r'){
            printf("\n");
        }
    }
}