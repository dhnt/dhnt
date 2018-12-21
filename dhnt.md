# The DHNT Language Specification

## Alphabet

DHNT only includes 26 characters in the hex range of [61-7a] from [UTF-8](https://www.utf8-chartable.de/unicode-utf8-table.pl) 

It is divided into two categories: vowel in the first column and consonant in the rest and divided into five vowel groups as arranged in the following alphabet chart.

|  |  |  |  |  |  | 
|--|--|--|--|--|--|
|a |b |c |d |  |  |
|e |f |g |h |  |  |
|i |j |k |l |m |n |
|o |p |q |r |s |t |
|u |v |w |x |y |z |


Each character only and alway has one sound. The sound of the consonants follow their leading vowel characters.

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

In writing, the vowel character following a consonant can be omitted if it is the same as the alphabet sound.

for example: dhnt is equivalent to dahenito in written form as d h n t is da he ni to respectively in the alphabet chart.
As you may have now noticed, d h n t are the last characters of the first four vowel groups in the alphabet chart.

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
The year of December 18 2018: 
aba ahe bajiahe or in contracted form: ab ah bjiah

Binary number or booleans are prefixed with bu.

|Boolean|Sound|Contracted|English|
|-------|-----|----------|-------|
|1      |bua  |bua       |true   |
|0      |buba |bub       |false  |


## Grammar

## Vocabulary

Word consists of syllables and stress always falls on the second last.

|Word    |English|
|--------|-------|
|u | kind
|bu | binary, boolean
|ju | decimal
|zu | collection, community

dahenito zu - the people who speak dhnt

loan word:

Any word in any other languages, natural or constructed, are considered valid word in dhnt after conversion into its 26 character alphabet through romanization or transliteration.