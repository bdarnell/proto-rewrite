#ifndef DESCRIPTORS_H
#define DESCRIPTORS_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// Converts a file within a FileDescriptorSet into a .proto file.
// The FileDescriptorSet is passed as an encoded blob; the return
// value is the contents of the .proto file as a null-terminated
// string (allocated with malloc; caller is responsible for freeing).
const char* decompile_proto(const char* descriptor_set_data,
    size_t descriptor_set_len, const char* filename);

#ifdef __cplusplus
}
#endif

#endif
