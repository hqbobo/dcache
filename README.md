
# Dcache  
  
## Description  
  
>distribute cache implement with redis. local memory is also used which can speed up get operation.<br>
main function as below

1. support redis and redis cluster mode  
2. support local cache. all set operation will remian a copy of data in memory(with timeout).
3. All dcache client will get a copy of data after set opertion and save in their own memory(with timeout).
4. All data has been marshel into josn, feel free to store struct stuff

## Functions  

> Create new picture with Configure
 
`pic := text2pic.NewTextPicture(text2pic.Configure{Width: 720, BgColor:text2pic.ColorWhite})`

> Add text line to picture with font, color and padding 

`pic.AddTextLine(" The Turkish lira plunged as much as 11% against the dollar", 13, f, text2pic.ColorBlue, text2pic.Padding{Left: 20, Right: 20, Bottom: 30})`

> Add picture  io.reader is required and padding as well

`pic.AddPictureLine(file, text2pic.Padding{Bottom: 20})`

> Draw it on io.writer. TypePng and TypeJpeg are supported

`pic.Draw(writer, text2pic.TypeJpeg)`

## Example  
> 
```
package main

import (
	"fmt"

)

func main() {

	
}
```