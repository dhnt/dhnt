# The DHNT Language Specification

## Vocabulary and executable syntax

**dhnt-lang defines vocabulary and canonical word formation; Bash# defines
executable grammar and syntax.** dhnt-lang is one language: its words are defined here, and its [syntax and grammar are defined by Bash#](https://github.com/bashsharp/bashsharp#grammar); the Grammar section below covers word formation (syllables, stress and contraction). The Bash# grammar is normative. For now, Bash# is the formal structure between agents: humans can read and write it, but it is not meant as a human-to-human tool. A formal structure for natural-language prose is deferred until dhnt-lang expands to communication between agents or robots and humans. Conventional spellings are accepted
alternatives when the existing dhnt word-formation, contraction, and loan-word
rules relate them to a canonical word. For example, `ls` expands to `liso`
by rule; no vocabulary entry needs to be registered for this pair.

The canonical word still follows dhnt's alphabet, syllable, and phonetic
rules. This is a distinction between canonical and alternative spellings,
not between written and spoken forms; either spelling can appear in text.
The rules apply productively to words, rather than enumerating every possible
word in every source language.

The [Bash# language](https://github.com/bashsharp/bashsharp) owns executable
syntax, including token boundaries, quoting, operators, declarations, and
reserved syntactic forms. Alternative spellings do not add grammar productions
or change how Bash# parses source. Bashy is a runtime/CLI consuming that
language, not a second vocabulary specification.

**Status:** the spelling model below formalizes the role of the existing
dhnt rules. The encoder already implements vowel insertion for lowercase
ASCII words; general Bash# spelling resolution is not introduced by this
revision. Existing encoder APIs and Skill-CNL serialization retain their behavior;
see [Compatibility and implementation status](#compatibility-and-implementation-status).

## Alphabet

DHNT includes 26 characters from [UTF-8](https://en.wikipedia.org/wiki/UTF-8) in the hex range [0061-007A]


a b c d e f g h i j k l m n o p q r s t u v w x y z

It is divided into two categories: vowels (5) in the first column and consonants (21) in the rest of the columns and arranged into five vowel group rows as shown in the following alphabet table.

|  |  |  |  |  |  | 
|--|--|--|--|--|--|
|a |b |c |d |  |  |
|e |f |g |h |  |  |
|i |j |k |l |m |n |
|o |p |q |r |s |t |
|u |v |w |x |y |z |


Each character always has one sound only; the sound of the consonants follows their leading vowel characters in the same group as shown in the following sound table.

|  |  |  |  |  |  | 
|--|---|---|---|---|---|
|a |ba |ca |da |||
|e |fe |ge |he |||
|i |ji |ki |li |mi |ni |
|o |po |qo |ro |so |to |
|u |vu |wu |xu |yu |zu |


The following shows the equivalent sound in English in capital letters in the second column.

|Character|Sound like in English|
|:-------:|:----------|
|a | Arch
|b | BAr
|c | TSAr
|d | DArt
||
|e | Else
|f | FEd
|g | GEt
|h | HEck
||
|i | It
|j | JIgsaw
|k | KIt
|l | LId
|m | MId
|n | NIl
||
|o | Odd
|p | POrt
|q | CHOcolate
|r | ROyal
|s | SOrt
|t | TOy
||
|u | wOO
|v | VOOdoo
|w | WOOd
|x | SHOOt
|y | YOUth
|z | ZOO


A syllable is composed of either one vowel on its own (a e i o u)
or one consonant followed by one vowel as shown in the following sound table.

|a  |e  |i  |o  |u  | 
|---|---|---|---|---|
|   |||||
|ba |be |bi |bo |bu |
|ca |ce |ci |co |cu | 
|da |de |di |do |du |
|   |  ||||
|fa |fe |fi |fo |fu |
|ga |ge |gi |go |gu |
|ha |he |hi |ho |hu |
|   |   |  |||
|ja |je |ji |jo |ju |
|ka |ke |ki |ko |ku |
|la |le |li |lo |lu |
|ma |me |mi |mo |mu |
|na |ne |ni |no |nu |
|   |   |   |   ||
|pa |pe |pi |po |pu | 
|qa |qe |qi |qo |qu |
|ra |re |ri |ro |ru |
|sa |se |si |so |su |
|ta |te |ti |to |tu |
|   |   |   |   |   |
|va |ve |vi |vo |vu |
|wa |we |wi |wo |wu |
|xa |xe |xi |xo |xu |
|ya |ye |yi |yo |yu |
|za |ze |zi |zo |zu |

There are 110 syllables in total.


## Numerals

Decimal numbers are prefixed with ju. 

|Digit|dhnt |English|
|-----|-----|-------|
|1    |jua  | one
|2    |jub  | two
|3    |juc  | three
|4    |jud  | four
|5    |jue  | five
|6    |juf  | six
|7    |jug  | seven
|8    |juh  | eight
|9    |jui  | nine
|0    |juj  | zero

The prefix ju may be dropped depending on the context.
The date of December 18 2018: 
aba ahe bajiahe or in contracted form: ab ah bjiah

Binary number or booleans are prefixed with bu.

|Boolean|dhnt |English|
|-------|-----|-------|
|1      |bua  |true, one
|0      |bub  |false, zero


Hexadecimal is prefixed with pu.

|Hex|dhnt |English|
|---|-----|-------|
|1  |pua  |one
|2  |pub  |two
|3  |puc  |three
|4  |pud  |four
|5  |pue  |five
|6  |puf  |six
|7  |pug  |seven
|8  |puh  |eight
|9  |pui  |nine
|A  |puj  |ten
|B  |puk  |eleven
|C  |pul  |twelve
|D  |pum  |thirteen
|E  |pun  |fourteen
|F  |puo  |fifteen
|0  |pup  |zero

Prefix pu may be omitted depending on the context.

<!-- ## Logic

_Unary_

|du|T |F |Note|
|--|--|--|--|
|a |F |T |negation
|b |T |F |identity
|c |T |T |always true
|d |F |F |never true


_Binary_

|pu|T.T |T.F |F.T |F.F |Note |
|--|----|----|----|----|-----|
|a |F |F |F |T |
|b |F |F |T |F |
|c |F |F |T |T |
|d |F |T |F |F |
|e |F |T |F |T |
|f |F |T |T |F |
|g |F |T |T |T |
|h |T |F |F |F |and
|i |T |F |F |T |if and only if
|j |T |F |T |F |
|k |T |F |T |T |if then
|l |T |T |F |F |
|m |T |T |F |T |
|n |T |T |T |F |or
|o |T |T |T |T |tautology
|p |F |F |F |F |contradiction


 -->

## Grammar

dhnt has 5 vowel characters and 21 consonant characters and a total of 110 syllables.

A word is a sequence of one or more syllables and stress always falls on the second last.

There are no clusters of vowels or consonants. In writing however, syllables may be contracted per the following rules:

- Consonant character y between two vowels can be omitted.
- Vowel character in a consonant-vowel syllable from the alphabet table can be omitted if it is followed by a consonant or is the last in a word.

For example:

The word _ayaya_ can be shortened to _aaa_

A vowel letter written after another vowel does not form a cluster or
diphthong with it: it is a syllable of its own. For example, _you_ is the two
syllables _yo·u_, stressed on _yo_ (the second last), unlike the
single-syllable English word; _aaa_ is the three syllables _a·a·a_.


_dhnt_ is equivalent to _dahenito_ as d h n t are da he ni to respectively in the alphabet table.

As you may have noticed now, d h n t are the last characters of the first four vowel group rows in the alphabet table - _the name of this language_.


## Vocabulary

### Canonical words and rule-derived alternative spellings

An alternative spelling is accepted by the applicable dhnt rules, not by
membership in a registry. Canonical forms continue to obey the existing
alphabet and word rules. Source-language import rules govern transliteration
or romanization where needed; contractions follow the existing contraction
rules. The examples below illustrate these rules, not a closed dictionary.

For lowercase ASCII words, the existing reference encoder expands a spelling
by scanning from left to right:

1. A vowel emits itself.
2. A consonant followed by a vowel emits that consonant and vowel, consuming
   both input characters.
3. A consonant followed by another consonant or the end of the word emits
   itself followed by its alphabet-row vowel.

Thus `l` receives `i`, and `s` receives `o`: `ls` becomes `liso`. An
already expanded `liso` stays `liso`. `bash` and its contraction `bsh` both
expand to `basohe`. No registration or choice of one preferred alias is
required for any of these spellings.

| Alternative spelling | Canonical expansion |
|---|---|
| `ls` | `liso` |
| `cd` | `cada` |
| `bash`, `bsh` | `basohe` |
| `hello`, `hllo` | `helilo` |
| `dhnt` | `dahenito` |

#### Equivalence and reversibility

Let `N` denote the existing lowercase ASCII word expansion above. For inputs
in that rule's domain, spelling equivalence is defined by equal expansions:

```text
x ~ y  iff  N(x) = N(y)
N("ls") = N("liso") = "liso"
N("bash") = N("bsh") = N("basohe") = "basohe"
N(N(x)) = N(x)
```

`N` is mathematical notation here; the existing implementation is
`EncodeWord`. This formalization does not claim that `EncodeWord` implements
every source-language import or contraction rule in this specification.

`liso ↔ ls` expresses spelling equivalence, not a globally invertible
one-to-one function. Each supported input has a deterministic expansion, but
several valid spellings can share it. The rules therefore do not select one
original or preferred conventional spelling from a canonical form alone.
A tool that must reproduce the original source must retain that spelling.
For other languages, use the applicable documented import rules; do not
invent a universal inverse, a translation, or a new transliteration rule.

#### Homonyms and context

Equal canonical spellings establish equivalence of form, not necessarily one
meaning across all source languages or contexts. Distinct words or meanings
may share a canonical form and be treated as homonyms. This is permitted,
just as in natural languages; canonicalization need not give every meaning a
globally unique spelling.

`ls` and `liso`, when referring to the listing command, are alternative
spellings of the same word, not two different meanings. If another word
shares that canonical form, its meaning can remain distinct. Context such as
the surrounding sentence, subject domain, source language, or syntactic role
can distinguish the intended sense. If context is insufficient, ambiguity
remains; a shared expansion does not justify guessing a meaning.

A glossary can record meanings or bindings for a particular application, but
an exhaustive glossary is not required to establish rule-derived spelling
equivalence. Homonyms are not grounds for rejecting valid spellings or
requiring new registrations. In executable source, Bash# and the runtime's
name-resolution rules determine bindings; spelling normalization alone does
not decide which command runs.

#### Token boundaries and execution

Word normalization and source-code rewriting are separate operations. Bash#
determines token boundaries and preserves its existing parsing and execution
semantics. This revision does not replace source tokens with their canonical
expansions or make arbitrary argument data into vocabulary tokens.

For example, `ls -l "My Files"` remains executable source with exactly those
spellings. The dhnt rules explain the relation between `ls` and `liso`;
they do not rewrite `-l`, `"My Files"`, `./ls`, `ReadMe`, or quoted text such
as `"你好"`. Lowercasing prescribed when importing a word into dhnt does not
authorize changing case-sensitive command names, variables or paths.

Rule-derived equivalence neither installs an executable named `liso` nor
makes the current shell dispatch it as `ls`. A future Bash# integration must
specify the token positions it handles and how it resolves executable
bindings when spellings share an expansion. Such application-specific
bindings are not a registry of all possible dhnt words.

### Vocabulary classes

*Reserved Word*

|dhnt    |English|
|--------|-------|
|u  | kind, classification
|azu | all, any
|bu | binary, boolean
|ju | decimal
|pu | hexadecimal, logic
|zu | collection, community, plural


dahenito zu - the people who speak dhnt

*Standard Word*

Standard words consist of [Unicode](https://www.unicode.org/charts/) hex encodings.

Examples:

|Script|Unicode |dhnt |English|
|------|--------|-----|-------|
|!     |21 |pubaa |exclamation
|#     |23 |pubc |hash tag
|$     |24 |pubd |dollar
|.     |2E |pubn |dot
|?     |3F |pucao |question
|@     |40 |pudp | at sign
||
|.com ||pubani comi | dot com

*Common Word*

Any words and symbols in dhnt alphabet that are accepted internationally.

[Top Level Domain](https://www.icann.org/resources/pages/tlds-2012-02-25-en)

Examples:

|TLD|dhnt|
|---|----|
|com|comi |
|org|oroge |


[ISO 3166](https://en.wikipedia.org/wiki/List_of_ISO_3166_country_codes)

Examples:

|Code|dhnt|English|
|----|----|-------|
|CN  |cani |China, Chinese |
|ES  |eso  |Spain, Spanish |
|FR  |fero |France, French |
|GB,UK  |geba, uki |Britain, United Kingdom, English |
|US | uso | America, American English |
|RU | ru |Russia, Russian |



*Loan Word*

Words in any other languages - programming, constructed, or natural - can be transformed through transcription, transliteration, or romanization and rewritten in dhnt alphabet.

_Programming Languages_

As most programming languages use the same 26 characters, they are considered dhnt dialects.

_Esperanto_

Esperanto can be imported after transliteration per the following:

|Esperanto|dhnt|
|---------|----|
|aA |a |
|bB |b |
|cC |c |
|ĉĈ |q |
|dD |d |
|eE |e |
|fF |f |
|gG |g |
|ĝĜ |g |
|hH |h |
|ĥĤ |h |
|iI |i |
|jJ |j |
|ĵĴ |j |
|kK |k |
|lL |l |
|mM |m |
|nN |n |
|oO |o |
|pP |p |
|rR |r |
|sS |s |
|ŝŜ |x |
|tT |t |
|uU |u |
|ŭŬ |w |
|vV |v |
|zZ |z |



_English_

English alphabet is the same as dhnt and its words can be imported as-is after converting upper case letters into lower cases.

The lowercased word is itself an accepted dhnt form. Its canonical expansion
follows by the dhnt word-formation rules (`ls` → `liso`, `bash` → `basohe`) and is used for
identity, hashing and cross-language matching; nobody is required to write
it. No pair registration is required. Importing a word does not authorize
lowercasing executable source identifiers.

_Chinese_

Chinese can be written in Pinyin - the standard Chinese romanization system and can be imported as-is.

_Latin and Cyrillic Script_

Languages in Latin and Cyrillic script can be imported after transliteration disregarding capitalization and diacritics per ISO-9.

## Compatibility and implementation status

This revision specifies rule-derived spelling equivalence without migrating existing data
or changing Go APIs, command dispatch, or content hashes. Skill-CNL input now
accepts alternative spellings while its output stays canonical.

- `EncodeWord` and `EncodePhrase` already generate full-form strings by rule,
  without a dictionary. `EncodeWord` still accepts lowercase ASCII input only;
  `EncodePhrase` still lowercases and normalizes whitespace. Neither proves
  unique meaning or serves as a source-code rewriting API. Language-specific
  import rules described elsewhere are not all implemented by these APIs.
- `IsCanonical` still validates full-form strings. In particular,
  `IsCanonical("ls")` remains false even though `ls` is a rule-derived
  alternative spelling of `liso`. Alternative-spelling acceptance and
  canonical-form validation are distinct operations.
- The existing [Skill-CNL specification](skills/SPEC.md) continues to define
  its own serialization grammar. `ParseDhnt` now accepts rule-derived
  alternative spellings on input: each keyword, name or atom token is read
  by its expansion `N(token)`, so `skill ls fin` reads as
  `sokilili liso fini`. A token whose expansion matches more than one entry
  in the same grammatical position is reported as ambiguous, never guessed.
  Output stays canonical: the AST stores only canonical forms, and its
  identifiers, `LineariseDhnt`/`ParseDhnt` roundtrip, execution bindings,
  and `Identity` hashes remain unchanged. That format is not Bash# source
  syntax.
- The existing `skills.Glossary` records application-specific meanings and
  labels for Skill-CNL. Its closed vocabulary does not define the set of all
  words admitted by dhnt's productive word rules. Its label normalization,
  duplicate checks, and the catalog's alias behavior remain unchanged.
- A vowel written after another vowel is a separate syllable, never a
  cluster, so the encoder's output for words such as `you` (`yo·u`) conforms
  to the [Grammar](#grammar) section; no encoder or hash change is involved.

Bash# token integration requires separate implementation work. An exhaustive vocabulary registry is not required or
planned by this spelling model. No new parser annotation notation, aliases
in executors, or speech interface is introduced here.
