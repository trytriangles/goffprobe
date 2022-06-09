# go-ffprobe

Go wrapper around ffprobe, providing a simplified breakdown of video and audio
file attributes.

```
>> info := NewVideoInfo("foo.mp4")
>> info.SimpleDescription
"h264+av1@2700000/aac@256000"
>> info.Pixels // Pixel count of highest-resolution video stream
414720
>> info.Duration // Seconds
618
>> info.HasMultipleVideo
true
>> info.AudioBitrate
256000
>> info.VideoBitrates
1800000, 900000
```

# License (MIT)

© 2022 Ryan Plant

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

The Software is provided "as is", without warranty of any kind, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no event shall the authors or copyright holders be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the Software or the use or other dealings in the Software.
