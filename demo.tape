Output demo.gif

Require echo
Require git
Require gittag

Set Shell "bash"
Set FontSize 20
Set Width 800
Set Height 600

Type "git log --oneline" Sleep 500ms Enter

Sleep 2s

Type "gittag patch" Sleep 500ms Enter

Sleep 2s

Type "touch foo && git add foo && git commit -qm foo" Sleep 500ms Enter

Type "gittag pre-minor alpha.1" Sleep 500ms Enter

Sleep 2s

Type "touch bar && git add bar && git commit -qm bar" Sleep 500ms Enter

Type "gittag pre alpha.2" Sleep 500ms Enter

Sleep 2s

Type "touch baz && git add baz && git commit -qm baz" Sleep 500ms Enter

Type "gittag release" Sleep 500ms Enter

Sleep 2s

Type "git log --oneline" Sleep 500ms Enter

Sleep 5s
