# CMENU

[NAME](#NAME)  
[SYNOPSIS](#SYNOPSIS)  
[DESCRIPTION](#DESCRIPTION)  
[EXAMPLE](#EXAMPLE)  
[SEE ALSO](#SEE%20ALSO)  
[AUTHOR](#AUTHOR)  

------------------------------------------------------------------------

## NAME <span id="NAME"></span>

cmenu − clipboard menu wrapper

## SYNOPSIS <span id="SYNOPSIS"></span>

**cmenu** *MENU* \[*FILE*\]

## DESCRIPTION <span id="DESCRIPTION"></span>

**cmenu** is a clipboard menu wrapper, originally designed to work with
*fzf*(1). **cmenu** wraps *MENU* to choose from JSON key−value entries
of type string in *FILE* or standard input. **cmenu** pipes all keys to
*MENU* which must then output a valid key. **cmenu** then outputs the
value associated with the key.

## EXAMPLE <span id="EXAMPLE"></span>

**\$ echo ’{"key":"value"}’ \| cmenu fzf**

## SEE ALSO <span id="SEE ALSO"></span>

***fzf***(1)

## AUTHOR <span id="AUTHOR"></span>

andrieee44 (andrieee44@gmail.com)

------------------------------------------------------------------------
