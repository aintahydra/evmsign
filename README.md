# evmsign
Convinently singing on multiple files utilizing the evmctl utility.

Usage:
$ ./evmsign -in=_filelist_ -key=_signkey_ -pdgree=_number_
 - _filelist_ := a text file contains a list of directories. Regular files
  under the directories are all signed
 - _signkey_ := RSA signing key (private key)
 - _number_ := a number of go routines that will sign on files in parallel

# NOTE
- You should install the "evmctl" utility (e.g., "apt-get install ima-evm-utils")
- Running "evmctl" may require the administrative privilege

# Example

Let's say we want to sign on files of a given project that has the following directory structure:
```
CERTAIN_DIR/myproject/
├── db-files
│   ├── a.db
│   └── b.db
├── mybin
└── mybin.policy
```

Write a filelist file, "target.txt"
(*I recommend not to use the tilde symbol inside this file, since you may need to run the program with the 'sudo' command. Better use the absolute paths.*):
```
CERTAIN_DIR/myproject
```

Then, run as follows:
```
$ sudo ./evmsign -in=./target.txt -key=./privkey.pem -pdgree=1
Signing with 1 routine(s)
Sign on:  /home/wshin/myproject/db-files/a.db
Sign on:  /home/wshin/myproject/db-files/b.db
Sign on:  /home/wshin/myproject/mybin
Sign on:  /home/wshin/myproject/mybin.policy
Done! It took  44.758064ms
```

For a reference, when it is run with 2 go routines:
```
sudo ./evmsign -in=./target.txt -key=./privkey.pem -pdgree=2
Signing with 2 routine(s)
...
Done! It took  27.154066ms
```

# Tips

## How to generate keys
A key pair can be created using the openssl package

### Create a configuration file

An example configuration file, "X509.genkey":
```
[ req ]
default_bits = 4096
distinguished_name = req_distinguished_name
prompt = no
string_mask = utf8only
x509_extensions = myexts

[ req_distinguished_name ]
O = aintahydra.github.io
CN = aintahydra
emailAddress = aintahydra@gmain.com

[ myexts ]
basicConstraints=critical,CA:FALSE
keyUsage=digitalSignature
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid
```
### Generate a key pair

```
$ openssl req -new -nodes -utf8 -sha256 -days 1000 -batch -x509 -config X509.genkey -outform DER -out x509_evm.der -keyout privkey.pem
Generating a 4096 bit RSA private key
.........................................................................................................++
..........................++
writing new private key to 'privkey.pem'
-----
```

### A simple signing test with the created key pair


Gerneate a sample file

```
$ echo "hello" > afile.txt
```

Sign on the file
```
sudo evmctl ima_sign --key privkey.pem afile.txt
```

Check if the sign is added
```
$  getfattr -m security -d afile.txt
# file: afile.txt
security.ima=0sAwIC...
```

Try to verify the signature (the generated certificate file should be placed in /etc/keys for the verification)
```
$ sudo mkdir /etc/keys
$ sudo cp x509_evm.der /etc/keys
$ evmctl ima_verify afile.txt
```

# References
- the EVMCTL man page
- https://wiki.gentoo.org/wiki/Signed_kernel_module_support

