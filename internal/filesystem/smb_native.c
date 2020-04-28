#include <errno.h>
#include <string.h>

#include "smb_native.h"

/* Struct definitions */

typedef struct user_data
{
     char *domain;
     char *password;
     char *username;
} user_data;

/* Function implementations */

void pipewerx_smb_auth_func (SMBCCTX *c, const char *srv, const char *shr, char *wg, int wglen, char *un, int unlen,
     char *pw, int pwlen)
{
     user_data *data = (user_data *) smbc_getOptionUserData(c);

     strncpy(un, data->username, unlen - 1);
     strncpy(pw, data->password, pwlen - 1);
}

int pipewerx_smb_close (SMBCCTX *context, SMBCFILE *file)
{
     return smbc_getFunctionClose(context)(context, file);
}

int pipewerx_smb_closedir (SMBCCTX *context, SMBCFILE *dir)
{
     return smbc_getFunctionClosedir(context)(context, dir);
}

SMBCCTX *pipewerx_smb_create_context (char *domain, char *username, char *password, bool enable_test_conditions)
{
     SMBCCTX *context;
     user_data *data;

     if (enable_test_conditions)
     {
          errno = ENOMEM;

          return NULL;
     }

     context = smbc_new_context();

     if (!context)
     {
          return NULL;
     }

     if (!smbc_init_context(context))
     {
          smbc_free_context(context, 1);

          return NULL;
     }

     data = malloc(sizeof(struct user_data));

     if (!data)
     {
          smbc_free_context(context, 1);

          return NULL;
     }

     data->domain = domain;
     data->password = password;
     data->username = username;

     smbc_setFunctionAuthDataWithContext(context, pipewerx_smb_auth_func);
     smbc_setOptionUserData(context, data);

     return context;
}

int pipewerx_smb_destroy_context (SMBCCTX *context, bool enable_test_conditions)
{
     user_data *data;

     if (enable_test_conditions)
     {
          errno = EBADF;

          return 1;
     }

     data = (user_data *) smbc_getOptionUserData(context);

     free(data->domain);
     free(data->password);
     free(data->username);
     free(data);

     return smbc_free_context(context, 1);
}

SMBCFILE *pipewerx_smb_open (SMBCCTX *context, const char *fname, int flags, mode_t mode)
{
     return smbc_getFunctionOpen(context)(context, fname, flags, mode);
}

SMBCFILE *pipewerx_smb_opendir (SMBCCTX *context, char *url)
{
     return smbc_getFunctionOpendir(context)(context, url);
}

ssize_t pipewerx_smb_read (SMBCCTX *context, SMBCFILE *file, void *buf, size_t count)
{
     return smbc_getFunctionRead(context)(context, file, buf, count);
}

const struct libsmb_file_info *pipewerx_smb_readdirplus2 (SMBCCTX *context, SMBCFILE *dir, struct stat *st,
     bool enable_test_conditions)
{
     if (enable_test_conditions)
     {
          errno = EBADF;

          return NULL;
     }

     return smbc_getFunctionReaddirPlus2(context)(context, dir, st);
}

int pipewerx_smb_stat (SMBCCTX *context, char *url, struct stat *st)
{
     return smbc_getFunctionStat(context)(context, url, st);
}
