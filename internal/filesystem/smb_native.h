#ifndef _SMB_NATIVE_H_
#define _SMB_NATIVE_H_

#include <libsmbclient.h>
#include <stdbool.h>
#include <stdlib.h>
#include <sys/stat.h>

/* Function definitions */

int pipewerx_smb_close (SMBCCTX *context, SMBCFILE *file);

int pipewerx_smb_closedir (SMBCCTX *context, SMBCFILE *dir);

SMBCCTX *pipewerx_smb_create_context (char *domain, char *username, char *password, bool enable_test_conditions);

int pipewerx_smb_destroy_context (SMBCCTX *context, bool enable_test_conditions);

SMBCFILE *pipewerx_smb_open (SMBCCTX *context, const char *fname, int flags, mode_t mode);

SMBCFILE *pipewerx_smb_opendir (SMBCCTX *context, char *url);

ssize_t pipewerx_smb_read (SMBCCTX *context, SMBCFILE *file, void *buf, size_t count);

const struct libsmb_file_info *pipewerx_smb_readdirplus2 (SMBCCTX *context, SMBCFILE *dir, struct stat *st,
     bool enable_test_conditions);

int pipewerx_smb_stat (SMBCCTX *context, char *url, struct stat *st);

#endif
