#include <unistd.h>
#include <sys/prctl.h>

int main(int argc, char *argv[]) {
    prctl(PR_SET_NAME, "qwertywm", 0, 0, 0);
    execvp("/usr/bin/river", argv);
    return 1;
}
