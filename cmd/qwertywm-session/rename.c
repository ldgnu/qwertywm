#include <sys/prctl.h>
#include <string.h>

__attribute__((constructor))
void rename_process() {
    prctl(PR_SET_NAME, "qwertywm", 0, 0, 0);
}
