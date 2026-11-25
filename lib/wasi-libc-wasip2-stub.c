// wasip2.c needs this symbol to keep the wasip2_component_type.o object.
// TinyGo has its own entry points and component embedding.
void __component_type_object_force_link_wasip2(void) {}
