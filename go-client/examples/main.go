package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/JanFalkin/tantivy-jpc/go-client/tantivy"
)

func DoRun() {
	tantivy.LibInit("debug")
	builder, err := tantivy.NewBuilder("/Users/mikhailyudin/GolandProjects/tantivy-jpc/go-client/examples/tmpdir")
	if err != nil {
		panic(err)
	}
	idxFieldTitle, err := builder.AddTextField("title", tantivy.TEXT, true, true, "", false)
	if err != nil {
		panic(err)
	}

	idxFieldBody, err := builder.AddTextField("body", tantivy.TEXT, true, true, "", false)
	if err != nil {
		panic(err)
	}

	idxMyIdId, err := builder.AddTextField("myId", tantivy.TEXT, true, true, "", false)
	if err != nil {
		panic(err)
	}

	idxFieldSpaceId, err := builder.AddTextField("spaceId", tantivy.TEXT, true, true, "", false)
	if err != nil {
		panic(err)
	}

	doc, err := builder.Build()
	if err != nil {
		panic(err)
	}

	idx, err := doc.CreateIndex()
	if err != nil {
		panic(err)
	}
	_, err = idx.SetMultiThreadExecutor(8)
	if err != nil {
		panic(err)
	}

	start := time.Now().UnixMilli()
	//var docs []SearchDoc
	startReportMemory()
	fmt.Println(idxFieldTitle, idxFieldBody, idxFieldSpaceId, idxMyIdId)
	idw, err := addDocs(doc, idxFieldTitle, idxFieldBody, idxFieldSpaceId, idxMyIdId, err, idx)
	//idw, err := add1DocMb(doc, idxFieldTitle, idxFieldBody, idxFieldSpaceId, idxFieldOrder, err, idx)

	_, err = idw.Commit()
	fmt.Println("time to add 100000 docs", time.Now().UnixMilli()-start)
	if err != nil {
		panic(err)
	}

	rb, err := idx.ReaderBuilder()
	if err != nil {
		panic(err)
	}

	qp, err := rb.Searcher()
	if err != nil {
		panic(err)
	}

	_, err = qp.ForIndex([]string{"title", "body"})
	if err != nil {
		panic(err)
	}

	searcher, err := qp.ParseQuery("february rates Indexx nov")
	if err != nil {
		panic(err)
	}

	var sr []map[string]interface{}

	s, err := searcher.Search(true, 100, 0, true)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal([]byte(s), &sr)
	if err != nil {
		panic(err)
	}
	fmt.Println("search result", sr[0]["doc"].(map[string]interface{})["myId"].([]interface{})[0])
	//if sr[0]["doc"].(map[string]interface{})["title"].([]interface{})[0] != "The Old Man and the Sea" {
	//	panic("expected value not received")
	//}
	//if err != nil {
	//	panic(err)
	//}
	//searcherAgain, err := qp.ParseQuery("order:222")
	//if err != nil {
	//	panic(err)
	//}
	//s, err = searcherAgain.Search(false, 0, 0, true)
	//if err != nil {
	//	panic(err)
	//}
	//err = json.Unmarshal([]byte(s), &sr)
	//if err != nil {
	//	panic(err)
	//}
	//
	//if sr[0]["doc"].(map[string]interface{})["title"].([]interface{})[0] != "Of Mice and Men" {
	//	panic("expected value not received")
	//}
	//
	//tantivy.ClearSession(builder.ID())
	fmt.Println("It worked!!!")

}

func addDocs(doc *tantivy.TDocument, idxFieldTitle int, idxFieldBody int, idxFieldSpaceId int, idxMyIdId int, err error, idx *tantivy.TIndex) (*tantivy.TIndexWriter, error) {
	for i := 1; i <= 1; i++ {
		for i := 1; i <= 100; i++ {
			{
				toAdd, err := doc.Create()
				if err != nil {
					panic(err)
				}
				doc.AddText(idxFieldTitle, getRandomString(), toAdd)
				doc.AddText(idxFieldBody, getRandomString(), toAdd)
				doc.AddText(idxFieldSpaceId, "spaceId", toAdd)
				doc.AddText(idxMyIdId, generateRandomString(8), toAdd)
			}
		}
	}

	idw, err := idx.CreateIndexWriter()
	if err != nil {
		panic(err)
	}

	maxJ := 100
	for i := 0; i <= 0; i++ {
		for j := 1; j <= maxJ; j++ {
			{
				_, err = idw.AddDocument(uint(i*maxJ + j))
				if err != nil {
					panic(err)
				}
			}
		}
	}
	return idw, err
}

func add1DocMb(doc *tantivy.TDocument, idxFieldTitle int, idxFieldBody int, idxFieldSpaceId int, idxFieldOrder int, err error, idx *tantivy.TIndex) (*tantivy.TIndexWriter, error) {
	for i := 1; i <= 1; i++ {
		for i := 1; i <= 1; i++ {
			{
				toAdd, err := doc.Create()
				if err != nil {
					panic(err)
				}
				doc.AddText(idxFieldTitle, "", toAdd)
				doc.AddText(idxFieldBody, getRandomStringMb(), toAdd)
				doc.AddText(idxFieldSpaceId, "spaceId", toAdd)
				doc.AddInt(idxFieldOrder, int64(i), toAdd)
			}
		}
	}

	idw, err := idx.CreateIndexWriter()
	if err != nil {
		panic(err)
	}

	for i := 1; i <= 1; i++ {
		for i := 1; i <= 1; i++ {
			{
				_, err = idw.AddDocument(uint(i))
				if err != nil {
					panic(err)
				}
			}
		}
	}
	return idw, err
}

func main() {
	DoRun()
}

func getRandomString() string {
	strLen := rand.Intn(len(english))
	wordArr := make([]string, strLen)
	for i := 0; i < strLen; i++ {
		wordArr[i] = english[rand.Intn(len(english))]
	}
	return strings.Join(wordArr, " ")
}

func getRandomStringMb() string {
	strLen := rand.Intn(len(english))
	wordArr := make([]string, 0)
	maxCount := 0
	for {
		nextWord := english[rand.Intn(strLen)]
		wordArr = append(wordArr, nextWord)
		maxCount += len(nextWord)
		if maxCount > 1024*1024 {
			break
		}
	}
	return strings.Join(wordArr, " ")
}

func generateRandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func startReportMemory() {
	if env := os.Getenv("ANYTYPE_REPORT_MEMORY"); env != "1" {
		go func() {
			var maxAlloc uint64
			var meanCPU float64
			var maxHeapObjects uint64
			var m runtime.MemStats
			times := 10 * 1
			for {
				runtime.ReadMemStats(&m)
				if maxAlloc < m.Alloc {
					maxAlloc = m.Alloc
				}
				if maxHeapObjects < m.HeapObjects {
					maxHeapObjects = m.HeapObjects
				}
				fmt.Println(
					map[string]uint64{
						"MaxAlloc":    uint64(float64(maxAlloc) / 1024 / 1024),
						"TotalAlloc":  uint64(float64(m.TotalAlloc) / 1024 / 1024),
						"Mallocs":     m.Mallocs,
						"Frees":       m.Frees,
						"MeanCpu":     uint64(meanCPU / float64(times)),
						"HeapObjects": maxHeapObjects,
					},
				)
				time.Sleep(time.Second)
			}
		}()
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var english = strings.Split(dict, "\n")

const dict = `item
international
center
ebay
must
store
travel
comments
made
development
report
off
member
details
line
terms
before
hotels
did
send
right
type
because
local
those
using
results
office
education
national
car
design
take
posted
internet
address
community
within
states
area
want
phone
dvd
shipping
reserved
subject
between
forum
family
l
long
based
w
code
show
o
even
black
check
special
prices
website
Indexx
being
women
much
sign
file
link
open
today
technology
south
case
project
same
pages
uk
version
section
own
found
sports
house
related
security
both
g
county
american
photo
game
members
power
while
care
network
down
computer
systems
three
total
place
end
following
download
h
him
without
per
access
think
north
resources
current
posts
big
media
law
control
water
history
pictures
size
art
personal
since
including
guide
shop
directory
board
location
change
white
text
small
rating
rate
government
children
during
usa
return
students
v
shopping
account
times
sites
level
digital
profile
previous
form
events
love
old
john
main
call
hours
image
department
title
description
non
k
y
insurance
another
why
shall
property
class
cd
still
money
quality
every
listing
content
country
private
little
visit
save
tools
low
reply
customer
december
compare
movies
include
college
value
article
york
man
card
jobs
provide
j
food
source
author
different
press
u
learn
sale
around
print
course
job
canada
process
teen
room
stock
training
too
credit
point
join
science
men
categories
advanced
west
sales
look
english
left
team
estate
box
conditions
select
windows
photos
gay
thread
week
category
note
live
large
gallery
table
register
however
june
october
november
market
library
really
action
start
series
model
features
air
industry
plan
human
provided
tv
yes
required
second
hot
accessories
cost
movie
forums
march
la
september
better
say
questions
july
yahoo
going
medical
test
friend
come
dec
server
pc
study
application
cart
staff
articles
san
feedback
again
play
looking
issues
april
never
users
complete
street
topic
comment
financial
things
working
against
standard
tax
person
below
mobile
less
got
blog
party
payment
equipment
login
student
let
programs
offers
legal
above
recent
park
stores
side
act
problem
red
give
memory
performance
social
q
august
quote
language
story
sell
options
experience
rates
create
key
body
young
america
important
field
few
east
paper
single
ii
age
activities
club
example
girls
additional
password
z
latest
something
road
gift
question
changes
night
ca
hard
texas
oct
pay
four
poker
status
browse
issue
range
building
seller
court
february
always
result
audio
light
write
war
nov
offer
blue
groups
al
easy
given
files
event
release
analysis
request
fax
china
making
picture
needs
possible
might
professional
yet
month
major
star
areas
future
space
committee
hand
sun
cards
problems
london
washington
meeting
rss
become
interest
id
child
keep
enter
california
share
similar
garden
schools
million
added
reference
companies
listed
baby
learning
energy
run
delivery
net
popular
term
film
stories
put
computers
journal
reports
co
try
welcome
central
images
president
notice
original
head
radio
until
cell
color
self
council
away
includes
track
australia
discussion
archive
once
others
entertainment
agreement
format
least
society
months
log
safety
friends
sure

faq
trade
edition
cars
messages
marketing
tell
further
updated
association
able
having
provides
david
fun
already
green
studies
close
common
drive
specific
several
gold
feb
living
sep
collection
called
short
arts
lot
ask
display
limited
powered
solutions
means
director
daily
beach
past
natural
whether
due
et
electronics
five
upon
period
planning
database
says
official
weather
mar
land
average
done
technical
window
france
pro
region
island
record
direct
microsoft
conference
environment
records
st
district
calendar
costs
style
url
front
statement
update
parts
aug
ever
downloads
early
miles
sound
resource
present
applications
either
ago
document
word
works
material
bill
apr
written
talk
federal
hosting
rules
final
adult
tickets
thing
centre
requirements
via
cheap
kids
finance
true
minutes
else
mark
third
rock
gifts
europe
reading
topics
bad
individual
tips
plus
auto
cover
usually
edit
together
videos
percent
fast
function
fact
unit
getting
global
tech
meet
far
economic
en
player
projects
lyrics
often
subscribe
submit
germany
amount
watch
included
feel
though
bank
risk
thanks
everything
deals
various
words
linux
jul
production
commercial
james
weight
town
heart
advertising
received
choose
treatment
newsletter
archives
points
knowledge
magazine
error
camera
jun
girl
currently
construction
toys
registered
clear
golf
receive
domain
methods
chapter
makes
protection
policies
loan
wide
beauty
manager
india
position
taken
sort
listings
models
michael
known
half
cases
step
engineering
florida
simple
quick
none
wireless
license
paul
friday
lake
whole
annual
published
later
basic
sony
shows
corporate
google
church
method
purchase
customers
active
response
practice
hardware
figure
materials
fire
holiday
chat
enough
designed
along
among
death
writing
speed
html
countries
loss
face
brand
discount
higher
effects
created
remember
standards
oil
bit
yellow
political
increase
advertise
kingdom
base
near
environmental
thought
stuff`
