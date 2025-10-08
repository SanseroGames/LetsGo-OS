#include <stdio.h>

int main(int argc, char **argv, char **envp)
{
  for (char **env = envp; *env != 0; env++)
  {
    char *thisEnv = *env;
    printf("%s\n", thisEnv);    
  }
  return 0;
}