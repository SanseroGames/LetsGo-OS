#include <stdio.h>

int main()
{
    printf("Start wasting time\n");
    int iters = 150000000;
    volatile int sink;
    do
    {
        sink = 0;
    } while (--iters > 0);
    printf("End wasting time\n");
}
