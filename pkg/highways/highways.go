package highways

/*
- Because county, state, and local highways are used as street names, these are not abbreviated.
- Please note that if the highway is a state highway, then the state is abbreviated following abbreviations in
Appendix D, but the word highway is not abbreviated.
- Note: When the name of a state is used as a portion of the Primary Street Name, developers SHOULD use
the standard two–letter abbreviation. However, when the state name is the complete Primary Street Name,
such as OKLAHOMA AVE, then the state name SHOULD be spelled out completely

  Example                               Project US@
COUNTY HIGHWAY 140             <->    COUNTY HIGHWAY 140
The word county is not abbreviated if part of a street name.

COUNTY HWY 60E                 <->    COUNTY HIGHWAY 60E
CNTY HWY 20                    <->    COUNTY HIGHWAY 20
Neither the word county nor the word highway should be abbreviated because they are
part of the street name.

COUNTY RD 441                  <->    COUNTY ROAD 441
Road is not abbreviated because it is part of the street name.

CR 1185                        <->    COUNTY ROAD 1185
CNTY RD 33                     <->    COUNTY ROAD 33
CA COUNTY RD 150               <->    CA COUNTY ROAD 150
Road is not abbreviated because it is part of the street name.

CALIFORNIA COUNTY ROAD 555     <->    CA COUNTY ROAD 555
EXPRESSWAY 55                  <->    EXPRESSWAY 55
FARM to MARKET 1200            <->    FM 1200
Farm to Market is always abbreviated to FM.

FM 187                         <->    FM 187

HWY FM 1320                    <->    FM 1320
It is incorrect to place the word Highway (or HWY) before FM

HWY 64                         <->    HIGHWAY 64
The word highway is not abbreviated because it is part of the street name.

HWY 11 BYPASS                  <->    HIGHWAY 11 BYP
The word bypass is abbreviated as a suffix, and not part of the street name.

HWY 66 FRONTAGE ROAD           <->    HIGHWAY 66 FRONTAGE RD
The word frontage is never abbreviated. Road is abbreviated as a suffix.

HIGHWAY 3 BYP ROAD             <->    HIGHWAY 3 BYPASS RD
The word bypass is not abbreviated because it is part of the street name.

I10                            <->    INTERSTATE 10
The word interstate is never abbreviated.

IH280                          <->    INTERSTATE 280
INTERSTATE HWY 680             <->    INTERSTATE 680

I 55 BYPASS                    <->    INTERSTATE 55 BYP
Interstate is never abbreviated. Bypass is abbreviated as a suffix.

I 26 BYP ROAD                  <->    INTERSTATE 26 BYPASS RD
Bypass is not abbreviated as it is part of the street name.

I 44 FRONTAGE ROAD             <->    INTERSTATE 44 FRONTAGE RD
Road is abbreviated as a suffix.

RD 5A                          <->    ROAD 5A
Road is not abbreviated if it is part of the street name.

RT 88                          <->    ROUTE 88
The word route is only abbreviated if it is a suffix, but not part of the street name.

RTE 95                         <->    ROUTE 95
RANCH RD 620                   <->    RANCH ROAD 620
ST HIGHWAY 303                 <->    STATE HIGHWAY 303
STATE HWY 60                   <->    STATE HIGHWAY 60
SR 220                         <->    STATE ROAD 220
ST RD 86                       <->    STATE ROAD 86
SR MM                          <->    STATE ROUTE MM
ST RT 175                      <->    STATE ROUTE 175
STATE RTE 260                  <->    STATE ROUTE 260
TOWNSHIP RD 20                 <->    TOWNSHIP ROAD 20
Road is not abbreviated as it is part
of the street name.

TSR 45                         <->    TOWNSHIP ROAD 45
The word township is never abbreviated.

US 41 SW                       <->    US HIGHWAY 41 SW
US HWY 44                      <->    US HIGHWAY 44
KENTUCKY 440                   <->    KY HIGHWAY 440
KENTUCKY HIGHWAY 189           <->    KY HIGHWAY 189

State names that are part of street names may be abbreviated following Appendix D.
KY 1207                        <->    KY HIGHWAY 1207
KY HWY 75                      <->    KY HIGHWAY 75
KY ST HWY 1                    <->    KY STATE HIGHWAY 1
The word state is not abbreviated if part of a street name.

KENTUCKY STATE HIGHWAY 625     <->   KY STATE HIGHWAY 625

*/
