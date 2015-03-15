#include <google/protobuf/descriptor.h>
#include <google/protobuf/descriptor.pb.h>
#include "descriptors.h"

inline bool ends_with(std::string const &value, std::string const &ending)
{
    if (ending.size() > value.size()) return false;
    return std::equal(ending.rbegin(), ending.rend(), value.rbegin());
}

const char* decompile_proto(const char* descriptor_set_data,
    size_t descriptor_set_len, const char* filename) {

    google::protobuf::FileDescriptorSet descriptor_set;
    descriptor_set.ParseFromArray(descriptor_set_data, descriptor_set_len);

    google::protobuf::DescriptorPool pool;

    for (auto file: descriptor_set.file()) {
        auto file_desc = pool.BuildFile(file);
        if (ends_with(filename, file.name())) {
            // DebugStringOptions is new in proto3.
            google::protobuf::DebugStringOptions dso;
            dso.include_comments = true;
            std::string output = file_desc->DebugStringWithOptions(dso);
            char* raw_output = static_cast<char*>(malloc(output.size()+1));
            memcpy(raw_output, output.data(), output.size());
            raw_output[output.size()] = 0;
            return raw_output;
        }
    }
    return NULL;
}
