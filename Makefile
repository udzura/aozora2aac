.PHONY: join clean convert_utf8
final.mp3:
	printf "file '%s'\n" out-*.mp3 > mylist.txt
	ffmpeg -f concat -safe 0 -i mylist.txt -c:a aac -b:a 128k final_audio.m4a
	rm -fv out-*.mp3 mylist.txt
join: final.mp3

clean:
	rm -fv out-*.mp3 mylist.txt final.mp3

input.txt:
	nkf -w $(INPUT_FILE) > input.txt
convert_utf8: input.txt
