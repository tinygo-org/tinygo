#include "main.h"

class FortyThree {
  int private_Member;
  public:
     FortyThree() {
	private_Member = 43;
     }
     int get() {
        return private_Member;
     }
};


extern "C" int fortythree() {
	FortyThree ft;

	return ft.get();
}

extern "C" int other() {
   return OTHER_CONSTANT;
}

