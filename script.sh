pip install  bs4==0.0.1 requests==2.5.0
pip install setuptools
python setup.py install
go build hasuratui.go scrape_hasura.go unqiue_list.go
sudo mv hasuratui /usr/bin/hasuratui
