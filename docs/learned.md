# What I Have Learned

In no particular order, these are the "new" things I have learned from working on this project. They are not strictly
"new", but something I found note-worthy, and... interesting. Probably I knew them before, but it is the first time I
actually put them into words.

## Binary Files Are Not That Scary

CS 101 taught me that, but working with binary files, or with a lower "layer" used to be a daunting task for me.
Creating a "Darkest Savior" really helped me in getting over that.

Everything is just... bytes in the end. Bytes are just... bits. 8 bits, to be precise. Bits are... ones (1s) and zeroes
(0s).

A text file with this content:

```
The quick brown fox jumps over the lazy dog.
```

In its simplest sense, the character `T` is represented by number `84`, which in turn represented by `0b1010100`. At the
lowest level, the computer is going to read the exact bits `1010100`, and be "smart" enough to turn it into a `T`.

Implementing Darkest Savior, I had to read a whole binary file, and put those bytes into meaningful structures. It makes
lower-level tasks' lost its "magic".

## Reality Has A Surprising Amount Of Details

I borrowed the phrase from another well-known article, as it was my exact thought after committing a fair amount of
code. "The Devil Is In The Details" is another right phrase for this case.

```
DSON -(1)-> JSON -(2)-> DSON

1: decoding
2: encoding
```

In its simplest sense, my code carry out the `1` and `2` processes.

- In `1`, I try to turn the binary file into a human-readable JSON file
- In `2`, I try to turn the JSON file into a valid DSON file

It is not a lot of work, right?

You have already known that the answer is "no".

In `1`, I have to do these jobs:

1. Read the header section
2. Read the meta 1 section
3. Read the meta 2 section
4. Read the actual data

I also expected it to be a straightforward implementation, but... reality has a surprising amount of details, you know.
Reading `4.` depends on the data of `2.` and `3.`, while reading `3.` and `2.` depends on `1.`'s. We have this
dependency graph:

```
1. ---> 2. ---> 4.
   \           ^
    \-> 3. ---/
```

Still, it was like that in the simplest sense, but more details creep in, and suddenly everything is so complex.

At first, I was not too content with the way that structure was implemented, but for now, try to see it from a "game
developer"'s lenses (even though I am not one) make me realize that this kind of binary structure is good for two
things:

- Gameplay: Darkest Dungeon is sort of an "unrevertable" game, where every action of you must have its consequences.
  Letting the player editing the save file breaks that.
- Programming: the game is going to persist the state of the game after every action of the player, which means
  continuously writing the whole file down is not a good strategy. This kind of proprietary format does the job well.

## Working With Go Is Not Too Horrible

Or it is still horrible, but I learned to appreciate the good parts of Go.

TODO
