# Sally Ride underway data feed

Use WICOR feed over udp.
Reference the `MetAcq.pdf` manual and `SR19D_temp.ACQ` configuration file
to interpret WICOR fields.
In cases where there are duplicate fields, the first is the primary and should
be used.
The exception are thermosalinograph fields, where the first set of values are from the bow.
The second set of values are from the main lab.
Then there are a third and fourth set of values which can be ignored (see ACQ file).

See the excel file in this directory for details about each feed in the example.

The feed frequency should be every 15 seconds.
If any values we select from a line are bad, throw out the entire line for simplicity.
