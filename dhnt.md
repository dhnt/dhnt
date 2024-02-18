# The DHNT Language Specification

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


_dhnt_ is equivalent to _dahenito_ as d h n t are da he ni to respectively in the alphabet table.

As you may have noticed now, d h n t are the last characters of the first four vowel group rows in the alphabet table - _the name of this language_.


## Vocabulary

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

_Chinese_

Chinese can be written in Pinyin - the standard Chinese romanization system and can be imported as-is.

_Latin and Cyrillic Script_

Languages in Latin and Cyrillic script can be imported after transliteration disregarding capitalization and diacritics per ISO-9.