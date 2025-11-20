# Overview

This doc describes the structure of the WAL to have durability.

## Structure

| Field        | Size (bytes) | Purpose                                                        |
| ------------ | ------------ | -------------------------------------------------------------- |
| Op Type      | 1            | What kind of operation (PUT vs DEL). 0x1 for PUT, 0x2 for DEL. |
| Key Length   | 4            | How many bytes are in the key                                  |
| Value Length | 4            | How many bytes are in the value? For deletes, this will be 0   |
| Key Bytes    | variable     | The actual bytes of the key.                                   |
| Value Bytes  | variable     | The actual bytes of the value. For deletes, this won't exist.  |
| CRC          | 4            | Checksum of all previous bytes                                 |
