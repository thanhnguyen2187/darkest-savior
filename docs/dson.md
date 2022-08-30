# DSON

**DSON** stands for **D**arkest Dungeon J**SON**, which is a proprietary data format created by Red Hook Studios.

Terminology and high level implementation were mostly based on documentation
by robojumper with his [DarkestDungeonSaveEditor](https://github.com/robojumper/DarkestDungeonSaveEditor/blob/master/docs/dson.md).

Decoding DSON code is a port from an [unfinished implementation in Janet](https://github.com/thanhnguyen2187/darkest-savior/blob/master/darkest-savior/dson.janet).

## Terminology

- `Header`: header of a DSON file; contains magic number and other general data
- `Meta1Block`: a block that contains metadata on each object in a DSON file
- `Meta2Block`: a block that contains metadata on each `Field` (JSON key-value pair) in a DSON file
- `Field`: a data unit that is equivalent to a key-value pair of JSON
- `Decode`: read bytes to raw data structures (`Header`, `Meta1Block`, `Meta2Block`, and `Field`)
- `Infer`: infer more meaningful data from raw data structures (result in `Meta1BlockInferences`, `Meta2BlockInferences`
  , etc.)

## DSON structure

A DSON file generally consists of four parts:

- `Header`
- `Meta1Blocks` 
- `Meta2Blocks`
- `Fields`

## Decoding DSON

Nothing too fancy is implemented. `Decode` functions expect a `BytesReader`, which is a wrapper around `bytes.Reader` with
some additional utilities like converting the read bytes to integer, or to a string.

As the manual mapping from `BytesReader` to raw data structures contains too much boilerplate code, an attempt of using
`ReadingInstruction` and `ExecuteInstructions` was formed.
