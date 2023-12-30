#define _GNU_SOURCE

#include <asm/ldt.h>
#include <errno.h>
#include <fcntl.h>
#include <linux/futex.h>
#include <linux/unistd.h>
#include <sched.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <sys/syscall.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/uio.h>
#include <sys/utsname.h>
#include <unistd.h>

int clone_test_to_test_exit(void *ptr)
{
	// SYS_EXIT
	printf("Test exit\n");
	syscall(SYS_exit, 42);
	return 0;
}

void main()
{
	// SYS_WRITE
	printf("Test write\n");

	// SYS_WRITEV
	printf("Test writeV\n");
	struct iovec iov[3];
	char text[][20] = {"hello ", "world!"};
	iov[0].iov_base = text[0];
	iov[1].iov_base = text[1];
	iov[2].iov_base = text[0];
	iov[0].iov_len = iov[1].iov_len = iov[2].iov_len = 6;
	writev(1, iov, 3);
	printf("\n");

	// SYS_SET_THREAD_AREA
	printf("Test set_thread_area\n");

	struct user_desc u_info = {
		.entry_number = -1,
		.base_addr = 0,
		.limit = 0xffffffff,
		.seg_32bit = 1,
		.contents = 0,
		.read_exec_only = 0,
		.limit_in_pages = 0,
		.seg_not_present = 0,
		.useable = 1};

	syscall(SYS_set_thread_area, (&u_info));

	printf(" Got entry number: %d\n", u_info.entry_number);

	// SYS_OPEN
	printf("Test open\n");
	open("/dev/null", O_CREAT, 0666);

	// SYS_OPENAT
	printf("Test openat\n");
	openat(42, "/dev/null", O_CREAT, 0666);

	// SYS_READ
	printf("Read test: ");
	char buf[2] = {0};
	fflush(stdout);
	read(0, &buf, 1);
	printf("\n You entered: %s\n", buf);

	// SYS_BRK
	printf("Test sbrk\n");
	printf(" Got %p\n", sbrk(0));

	// SYS_MMAP2
	printf("mmap test\n");
#define map_length 0x8000
	void *addr = mmap(NULL, map_length, PROT_READ, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
	printf(" Got addr: %p\n", addr);

	// SYS_MINCORE
	printf("Test mincore\n");
	char vec[(map_length + 0x1000 - 1) / 0x1000];
	mincore(addr, map_length, vec);
	for (int i = 0; i < (map_length + 0x1000 - 1) / 0x1000; i++)
	{
		printf("%c ", vec[1] + 0x30);
	}
	printf("\n");

	// SYS_MUNMAP
	printf("munmap test\n");
	munmap(addr, map_length);

	// SYS_GETTID
	pid_t tid = gettid();
	printf("test get_tid: %d\n", tid);

	// SYS_GETPID
	pid_t pid = getpid();
	printf("test get_pid: %d\n", pid);

	// SYS_FUTEX
	printf("Test Futex\n");
	// FUTEX_WAIT
	printf(" Futex wait\n");
	uint32_t uaddr;
	syscall(SYS_futex, &uaddr, FUTEX_WAIT | FUTEX_PRIVATE_FLAG, 42, NULL, NULL, 0);

	// FUTEX_WAKE
	printf(" Futex wake\n");
	syscall(SYS_futex, &uaddr, FUTEX_WAKE | FUTEX_PRIVATE_FLAG, 42, NULL, NULL, 0);

	// SYS_CLONE
	printf("Test clone\n");
	void *new_stack = malloc(0x4000);
	clone(clone_test_to_test_exit, new_stack, CLONE_THREAD, NULL);

	// SYS_UNAME
	printf("Test uname\n");
	struct utsname utsname;
	uname(&utsname);
	printf(" sysname: %s\n", utsname.sysname);
	printf(" nodename: %s\n", utsname.nodename);
	printf(" release: %s\n", utsname.release);
	printf(" version: %s\n", utsname.version);
	printf(" machine: %s\n", utsname.machine);
	printf(" domainname: %s\n", utsname.domainname);

	// SYS_SCHED_YIELD
	printf("Test sched yield\n");
	sched_yield();

	// statx syscall
	// use statx tool instead. Will be called as test program for execve

	// SYS_EXECVE
	// This behaviour is not the same on Let's Go OS and Linux!
	// Linux replaces the process.
	// Let's Go spawns a new process and keeps the old one running.
	printf("Test execve\n");
	// program that exists
	char *args[] = {"ls", "-l", "-a", (char *)0};
	char *env_args[] = {(char *)0};
	int ret = execve("/usr/statx", args, env_args);
	printf(" execve1: %x (%d)\n", ret, errno);
	// program that does not exist
	ret = execve("/this-program-does-not-exist", args, env_args);
	printf(" execve2: %x (%d)\n", ret, errno);

	// SYS_EXIT_GROUP
	printf("Test exit group\n");
	syscall(SYS_exit_group, 0);
}
