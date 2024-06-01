#include <unistd.h>
#include <sys/reboot.h>

int main()
{
    reboot(RB_POWER_OFF);
}