# -*- coding: utf-8 -*-

"""
A text interface to Taiga.
"""

from __future__ import print_function
import sys

from setuptools import setup, find_packages

REQUIREMENTS = [
    "bs4==0.0.1",
    "requests==2.5.0",
    "lxml==3.7.3 "
]


setup(name="scrape",
      version="1.0",
      description="a text scrape for website",
      packages=find_packages(),
      entry_points={
          "console_scripts": ["hscrape = hscrape.scrape:main"]
      },
      classifiers=[],
      install_requires=REQUIREMENTS,)
