join:
	printf "file '%s'\n" out-*.mp3 > mylist.txt
	ffmpeg -f concat -safe 0 -i mylist.txt -c copy final.mp3
	rm -fv out-*.mp3 mylist.txt

clean:
	rm -fv out-*.mp3 mylist.txt final.mp3