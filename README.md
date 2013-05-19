Atomic
======
Atomic is a househould reactor; (the beginning of!) a project to automate my 
flat.  Currently, it provides an interface for sending wake-on-lan (WOL) 
packets; the idea being externally-driven events (say, a smartphone entering 
home wifi range) can wake up our computers.

Over time, this will evolve to incorporate control of LimitlessLED bulbs, access
to various sensors, etc.

This Is
-------
My first project in Google Go.  It's very possible I've made bad/non-idiomatic
code.  Pull requests and bugs are welcome!

Acceptance Testing
------------------
Somewhere further down the line, I'll start writing unit tests.  For now, there's
some very broad acceptance tests, written using Python's [behave], an 
implementation of Cucumber.  Simply type

    behave

At your prompt (once you've setup the environment!) to run the acceptance tests.

The acceptance tests are housed in a [separate project].

[behave]: http://pythonhosted.org/behave/
[separate project]: http://github.com/frio/atomic-tests

Licensing
---------
Freely available for all and sundry under the MIT license!
