Implements a good enough JSON parser that:

1. Streams from a reader, yielding one token at a time
2. Handles comments (by stripping them out)
3. Exposes a lowish level API needed for deserializing polymorphic content w/o any

The generated OOTR files, and OOTR settings and spoilers can have comments and
polymorphic content in them and there's not a lot of choice on the shelf for
handling all all of the above. 

This implementation is very lax in syntax by allowing trailing commas and it
definitely does not have proper number and string deserialization. Don't let
it get wet, don't expose it to direct sunlight, and don't feed it
untrusted/adversial content.
