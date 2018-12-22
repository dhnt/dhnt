# The DHNT Language Specification

## Alphabet

DHNT includes 26 characters from [UTF-8](https://www.utf8-chartable.de/unicode-utf8-table.pl) in the hex range [61-7a]


a b c d e f g h i j k l m n o p q r s t u v w x y z

It is divided into two categories: vowels (5) in the first column and consonants (21) in the rest and arranged into five vowel groups as shown in the following alphabet table.

|  |  |  |  |  |  | 
|--|--|--|--|--|--|
|a |b |c |d |  |  |
|e |f |g |h |  |  |
|i |j |k |l |m |n |
|o |p |q |r |s |t |
|u |v |w |x |y |z |


Each character only and always has one sound; the sound of the consonants follows their leading vowel characters in the same group as shown in the following sound table.

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

|Digit|Sound|Contracted|
|-----|-----|----------|
|1    |jua  | jua
|2    |juba | jub
|3    |juca | juc
|4    |juda | jud
|5    |jue  | jue
|6    |jufe | juf
|7    |juge | jug
|8    |juhe | juh
|9    |jui  | jui
|0    |juji | juj

The prefix ju may be dropped depending on the context.
The date of December 18 2018: 
aba ahe bajiahe or in contracted form: ab ah bjiah

Binary number or booleans are prefixed with bu.

|Boolean|Sound|Contracted|English|
|-------|-----|----------|-------|
|1      |bua  |bua       |true   |
|0      |buba |bub       |false  |


## Grammar

dhnt has 5 vowel characters and 21 consonant characters and a total of 110 syllables.

A word is a sequence of one or more syllables and stress always falls on the second last.

There are no clusters of vowels or consonants. In writing however, syllables can be contracted as follows:

- Consonant character y can be omitted if the consecutive syllables have the same vowel.
- Vowel character can be omitted if the syllable is the same as the alphabet sound.

For example:

The word _ayaya_ can be shortened to _aaa_


_dhnt_ is equivalent to _dahenito_ as d h n t are da he ni to respectively in the alphabet sound table.

As you may have noticed now, d h n t are the last characters of the first four vowel groups - _the name of this language_.


## Vocabulary

*Reserved Word*

|Word    |English|
|--------|-------|
|u | kind
|bu | binary, boolean
|ju | decimal
|zu | collection, community

dahenito zu - the people who speak dhnt

*Loan Word*

Words in any other languages - programming, constructed, or natural - can be transformed through transcription, transliteration, or romanization and rewritten in dhnt alphabet.
