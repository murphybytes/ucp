#include <sys/types.h>
#include <security/pam_appl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>


int conv_func(int msgs, const struct pam_message **msg, struct pam_response **resp, void *p ){

  struct pam_response *aresp;

  if(msgs == 1){
      if((aresp = calloc(msgs, sizeof *aresp)) == NULL){
        return PAM_BUF_ERR;
      }

      const char *pwd = (const char*)p;

      aresp[0].resp = malloc(strlen(pwd)+1);
      strcpy(aresp[0].resp,pwd);
      *resp = aresp;
  }

  return PAM_SUCCESS;
}

int authorize_user(const char* user, const char* pwd ) {

  struct pam_conv pamc;
  pamc.conv = &conv_func;
  pamc.appdata_ptr = (void*)pwd;

  pam_handle_t *pamh=NULL;
  int retval;

  retval = pam_start("chkpasswd", user, &pamc, &pamh);

  if( retval == PAM_SUCCESS ) {
    retval = pam_authenticate(pamh, 0);
  }

  pam_end(pamh, retval);

  return retval;
}
