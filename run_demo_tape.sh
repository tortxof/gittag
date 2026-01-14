cd $(mktemp -dp .)
git init
touch main
git add main
git commit -m 'initial commit'
gittag init
echo foo > main
git add main
git commit -m 'second commit'
vhs ../demo.tape
mv demo.gif ..
rm -rf $(pwd)
