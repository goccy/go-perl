/* go-perl native XS SDK -- the PL_ppaddr proxy table. Every slot is a
 * distinct stub so a call through slot N carries the op type N into the
 * generic dispatcher (run_original_op(type) calls the saved slot for that
 * type, and the dispatcher must know which). GENERATED; do not edit.
 *
 * Requires: OP, pTHX, and goperl_pp_dispatch(pTHX_ int) declared. */
#ifndef GOPERL_PPSLOTS_H
#define GOPERL_PPSLOTS_H

static OP *goperl_pp_slot_0(pTHX) { return goperl_pp_dispatch(aTHX_ 0); }
static OP *goperl_pp_slot_1(pTHX) { return goperl_pp_dispatch(aTHX_ 1); }
static OP *goperl_pp_slot_2(pTHX) { return goperl_pp_dispatch(aTHX_ 2); }
static OP *goperl_pp_slot_3(pTHX) { return goperl_pp_dispatch(aTHX_ 3); }
static OP *goperl_pp_slot_4(pTHX) { return goperl_pp_dispatch(aTHX_ 4); }
static OP *goperl_pp_slot_5(pTHX) { return goperl_pp_dispatch(aTHX_ 5); }
static OP *goperl_pp_slot_6(pTHX) { return goperl_pp_dispatch(aTHX_ 6); }
static OP *goperl_pp_slot_7(pTHX) { return goperl_pp_dispatch(aTHX_ 7); }
static OP *goperl_pp_slot_8(pTHX) { return goperl_pp_dispatch(aTHX_ 8); }
static OP *goperl_pp_slot_9(pTHX) { return goperl_pp_dispatch(aTHX_ 9); }
static OP *goperl_pp_slot_10(pTHX) { return goperl_pp_dispatch(aTHX_ 10); }
static OP *goperl_pp_slot_11(pTHX) { return goperl_pp_dispatch(aTHX_ 11); }
static OP *goperl_pp_slot_12(pTHX) { return goperl_pp_dispatch(aTHX_ 12); }
static OP *goperl_pp_slot_13(pTHX) { return goperl_pp_dispatch(aTHX_ 13); }
static OP *goperl_pp_slot_14(pTHX) { return goperl_pp_dispatch(aTHX_ 14); }
static OP *goperl_pp_slot_15(pTHX) { return goperl_pp_dispatch(aTHX_ 15); }
static OP *goperl_pp_slot_16(pTHX) { return goperl_pp_dispatch(aTHX_ 16); }
static OP *goperl_pp_slot_17(pTHX) { return goperl_pp_dispatch(aTHX_ 17); }
static OP *goperl_pp_slot_18(pTHX) { return goperl_pp_dispatch(aTHX_ 18); }
static OP *goperl_pp_slot_19(pTHX) { return goperl_pp_dispatch(aTHX_ 19); }
static OP *goperl_pp_slot_20(pTHX) { return goperl_pp_dispatch(aTHX_ 20); }
static OP *goperl_pp_slot_21(pTHX) { return goperl_pp_dispatch(aTHX_ 21); }
static OP *goperl_pp_slot_22(pTHX) { return goperl_pp_dispatch(aTHX_ 22); }
static OP *goperl_pp_slot_23(pTHX) { return goperl_pp_dispatch(aTHX_ 23); }
static OP *goperl_pp_slot_24(pTHX) { return goperl_pp_dispatch(aTHX_ 24); }
static OP *goperl_pp_slot_25(pTHX) { return goperl_pp_dispatch(aTHX_ 25); }
static OP *goperl_pp_slot_26(pTHX) { return goperl_pp_dispatch(aTHX_ 26); }
static OP *goperl_pp_slot_27(pTHX) { return goperl_pp_dispatch(aTHX_ 27); }
static OP *goperl_pp_slot_28(pTHX) { return goperl_pp_dispatch(aTHX_ 28); }
static OP *goperl_pp_slot_29(pTHX) { return goperl_pp_dispatch(aTHX_ 29); }
static OP *goperl_pp_slot_30(pTHX) { return goperl_pp_dispatch(aTHX_ 30); }
static OP *goperl_pp_slot_31(pTHX) { return goperl_pp_dispatch(aTHX_ 31); }
static OP *goperl_pp_slot_32(pTHX) { return goperl_pp_dispatch(aTHX_ 32); }
static OP *goperl_pp_slot_33(pTHX) { return goperl_pp_dispatch(aTHX_ 33); }
static OP *goperl_pp_slot_34(pTHX) { return goperl_pp_dispatch(aTHX_ 34); }
static OP *goperl_pp_slot_35(pTHX) { return goperl_pp_dispatch(aTHX_ 35); }
static OP *goperl_pp_slot_36(pTHX) { return goperl_pp_dispatch(aTHX_ 36); }
static OP *goperl_pp_slot_37(pTHX) { return goperl_pp_dispatch(aTHX_ 37); }
static OP *goperl_pp_slot_38(pTHX) { return goperl_pp_dispatch(aTHX_ 38); }
static OP *goperl_pp_slot_39(pTHX) { return goperl_pp_dispatch(aTHX_ 39); }
static OP *goperl_pp_slot_40(pTHX) { return goperl_pp_dispatch(aTHX_ 40); }
static OP *goperl_pp_slot_41(pTHX) { return goperl_pp_dispatch(aTHX_ 41); }
static OP *goperl_pp_slot_42(pTHX) { return goperl_pp_dispatch(aTHX_ 42); }
static OP *goperl_pp_slot_43(pTHX) { return goperl_pp_dispatch(aTHX_ 43); }
static OP *goperl_pp_slot_44(pTHX) { return goperl_pp_dispatch(aTHX_ 44); }
static OP *goperl_pp_slot_45(pTHX) { return goperl_pp_dispatch(aTHX_ 45); }
static OP *goperl_pp_slot_46(pTHX) { return goperl_pp_dispatch(aTHX_ 46); }
static OP *goperl_pp_slot_47(pTHX) { return goperl_pp_dispatch(aTHX_ 47); }
static OP *goperl_pp_slot_48(pTHX) { return goperl_pp_dispatch(aTHX_ 48); }
static OP *goperl_pp_slot_49(pTHX) { return goperl_pp_dispatch(aTHX_ 49); }
static OP *goperl_pp_slot_50(pTHX) { return goperl_pp_dispatch(aTHX_ 50); }
static OP *goperl_pp_slot_51(pTHX) { return goperl_pp_dispatch(aTHX_ 51); }
static OP *goperl_pp_slot_52(pTHX) { return goperl_pp_dispatch(aTHX_ 52); }
static OP *goperl_pp_slot_53(pTHX) { return goperl_pp_dispatch(aTHX_ 53); }
static OP *goperl_pp_slot_54(pTHX) { return goperl_pp_dispatch(aTHX_ 54); }
static OP *goperl_pp_slot_55(pTHX) { return goperl_pp_dispatch(aTHX_ 55); }
static OP *goperl_pp_slot_56(pTHX) { return goperl_pp_dispatch(aTHX_ 56); }
static OP *goperl_pp_slot_57(pTHX) { return goperl_pp_dispatch(aTHX_ 57); }
static OP *goperl_pp_slot_58(pTHX) { return goperl_pp_dispatch(aTHX_ 58); }
static OP *goperl_pp_slot_59(pTHX) { return goperl_pp_dispatch(aTHX_ 59); }
static OP *goperl_pp_slot_60(pTHX) { return goperl_pp_dispatch(aTHX_ 60); }
static OP *goperl_pp_slot_61(pTHX) { return goperl_pp_dispatch(aTHX_ 61); }
static OP *goperl_pp_slot_62(pTHX) { return goperl_pp_dispatch(aTHX_ 62); }
static OP *goperl_pp_slot_63(pTHX) { return goperl_pp_dispatch(aTHX_ 63); }
static OP *goperl_pp_slot_64(pTHX) { return goperl_pp_dispatch(aTHX_ 64); }
static OP *goperl_pp_slot_65(pTHX) { return goperl_pp_dispatch(aTHX_ 65); }
static OP *goperl_pp_slot_66(pTHX) { return goperl_pp_dispatch(aTHX_ 66); }
static OP *goperl_pp_slot_67(pTHX) { return goperl_pp_dispatch(aTHX_ 67); }
static OP *goperl_pp_slot_68(pTHX) { return goperl_pp_dispatch(aTHX_ 68); }
static OP *goperl_pp_slot_69(pTHX) { return goperl_pp_dispatch(aTHX_ 69); }
static OP *goperl_pp_slot_70(pTHX) { return goperl_pp_dispatch(aTHX_ 70); }
static OP *goperl_pp_slot_71(pTHX) { return goperl_pp_dispatch(aTHX_ 71); }
static OP *goperl_pp_slot_72(pTHX) { return goperl_pp_dispatch(aTHX_ 72); }
static OP *goperl_pp_slot_73(pTHX) { return goperl_pp_dispatch(aTHX_ 73); }
static OP *goperl_pp_slot_74(pTHX) { return goperl_pp_dispatch(aTHX_ 74); }
static OP *goperl_pp_slot_75(pTHX) { return goperl_pp_dispatch(aTHX_ 75); }
static OP *goperl_pp_slot_76(pTHX) { return goperl_pp_dispatch(aTHX_ 76); }
static OP *goperl_pp_slot_77(pTHX) { return goperl_pp_dispatch(aTHX_ 77); }
static OP *goperl_pp_slot_78(pTHX) { return goperl_pp_dispatch(aTHX_ 78); }
static OP *goperl_pp_slot_79(pTHX) { return goperl_pp_dispatch(aTHX_ 79); }
static OP *goperl_pp_slot_80(pTHX) { return goperl_pp_dispatch(aTHX_ 80); }
static OP *goperl_pp_slot_81(pTHX) { return goperl_pp_dispatch(aTHX_ 81); }
static OP *goperl_pp_slot_82(pTHX) { return goperl_pp_dispatch(aTHX_ 82); }
static OP *goperl_pp_slot_83(pTHX) { return goperl_pp_dispatch(aTHX_ 83); }
static OP *goperl_pp_slot_84(pTHX) { return goperl_pp_dispatch(aTHX_ 84); }
static OP *goperl_pp_slot_85(pTHX) { return goperl_pp_dispatch(aTHX_ 85); }
static OP *goperl_pp_slot_86(pTHX) { return goperl_pp_dispatch(aTHX_ 86); }
static OP *goperl_pp_slot_87(pTHX) { return goperl_pp_dispatch(aTHX_ 87); }
static OP *goperl_pp_slot_88(pTHX) { return goperl_pp_dispatch(aTHX_ 88); }
static OP *goperl_pp_slot_89(pTHX) { return goperl_pp_dispatch(aTHX_ 89); }
static OP *goperl_pp_slot_90(pTHX) { return goperl_pp_dispatch(aTHX_ 90); }
static OP *goperl_pp_slot_91(pTHX) { return goperl_pp_dispatch(aTHX_ 91); }
static OP *goperl_pp_slot_92(pTHX) { return goperl_pp_dispatch(aTHX_ 92); }
static OP *goperl_pp_slot_93(pTHX) { return goperl_pp_dispatch(aTHX_ 93); }
static OP *goperl_pp_slot_94(pTHX) { return goperl_pp_dispatch(aTHX_ 94); }
static OP *goperl_pp_slot_95(pTHX) { return goperl_pp_dispatch(aTHX_ 95); }
static OP *goperl_pp_slot_96(pTHX) { return goperl_pp_dispatch(aTHX_ 96); }
static OP *goperl_pp_slot_97(pTHX) { return goperl_pp_dispatch(aTHX_ 97); }
static OP *goperl_pp_slot_98(pTHX) { return goperl_pp_dispatch(aTHX_ 98); }
static OP *goperl_pp_slot_99(pTHX) { return goperl_pp_dispatch(aTHX_ 99); }
static OP *goperl_pp_slot_100(pTHX) { return goperl_pp_dispatch(aTHX_ 100); }
static OP *goperl_pp_slot_101(pTHX) { return goperl_pp_dispatch(aTHX_ 101); }
static OP *goperl_pp_slot_102(pTHX) { return goperl_pp_dispatch(aTHX_ 102); }
static OP *goperl_pp_slot_103(pTHX) { return goperl_pp_dispatch(aTHX_ 103); }
static OP *goperl_pp_slot_104(pTHX) { return goperl_pp_dispatch(aTHX_ 104); }
static OP *goperl_pp_slot_105(pTHX) { return goperl_pp_dispatch(aTHX_ 105); }
static OP *goperl_pp_slot_106(pTHX) { return goperl_pp_dispatch(aTHX_ 106); }
static OP *goperl_pp_slot_107(pTHX) { return goperl_pp_dispatch(aTHX_ 107); }
static OP *goperl_pp_slot_108(pTHX) { return goperl_pp_dispatch(aTHX_ 108); }
static OP *goperl_pp_slot_109(pTHX) { return goperl_pp_dispatch(aTHX_ 109); }
static OP *goperl_pp_slot_110(pTHX) { return goperl_pp_dispatch(aTHX_ 110); }
static OP *goperl_pp_slot_111(pTHX) { return goperl_pp_dispatch(aTHX_ 111); }
static OP *goperl_pp_slot_112(pTHX) { return goperl_pp_dispatch(aTHX_ 112); }
static OP *goperl_pp_slot_113(pTHX) { return goperl_pp_dispatch(aTHX_ 113); }
static OP *goperl_pp_slot_114(pTHX) { return goperl_pp_dispatch(aTHX_ 114); }
static OP *goperl_pp_slot_115(pTHX) { return goperl_pp_dispatch(aTHX_ 115); }
static OP *goperl_pp_slot_116(pTHX) { return goperl_pp_dispatch(aTHX_ 116); }
static OP *goperl_pp_slot_117(pTHX) { return goperl_pp_dispatch(aTHX_ 117); }
static OP *goperl_pp_slot_118(pTHX) { return goperl_pp_dispatch(aTHX_ 118); }
static OP *goperl_pp_slot_119(pTHX) { return goperl_pp_dispatch(aTHX_ 119); }
static OP *goperl_pp_slot_120(pTHX) { return goperl_pp_dispatch(aTHX_ 120); }
static OP *goperl_pp_slot_121(pTHX) { return goperl_pp_dispatch(aTHX_ 121); }
static OP *goperl_pp_slot_122(pTHX) { return goperl_pp_dispatch(aTHX_ 122); }
static OP *goperl_pp_slot_123(pTHX) { return goperl_pp_dispatch(aTHX_ 123); }
static OP *goperl_pp_slot_124(pTHX) { return goperl_pp_dispatch(aTHX_ 124); }
static OP *goperl_pp_slot_125(pTHX) { return goperl_pp_dispatch(aTHX_ 125); }
static OP *goperl_pp_slot_126(pTHX) { return goperl_pp_dispatch(aTHX_ 126); }
static OP *goperl_pp_slot_127(pTHX) { return goperl_pp_dispatch(aTHX_ 127); }
static OP *goperl_pp_slot_128(pTHX) { return goperl_pp_dispatch(aTHX_ 128); }
static OP *goperl_pp_slot_129(pTHX) { return goperl_pp_dispatch(aTHX_ 129); }
static OP *goperl_pp_slot_130(pTHX) { return goperl_pp_dispatch(aTHX_ 130); }
static OP *goperl_pp_slot_131(pTHX) { return goperl_pp_dispatch(aTHX_ 131); }
static OP *goperl_pp_slot_132(pTHX) { return goperl_pp_dispatch(aTHX_ 132); }
static OP *goperl_pp_slot_133(pTHX) { return goperl_pp_dispatch(aTHX_ 133); }
static OP *goperl_pp_slot_134(pTHX) { return goperl_pp_dispatch(aTHX_ 134); }
static OP *goperl_pp_slot_135(pTHX) { return goperl_pp_dispatch(aTHX_ 135); }
static OP *goperl_pp_slot_136(pTHX) { return goperl_pp_dispatch(aTHX_ 136); }
static OP *goperl_pp_slot_137(pTHX) { return goperl_pp_dispatch(aTHX_ 137); }
static OP *goperl_pp_slot_138(pTHX) { return goperl_pp_dispatch(aTHX_ 138); }
static OP *goperl_pp_slot_139(pTHX) { return goperl_pp_dispatch(aTHX_ 139); }
static OP *goperl_pp_slot_140(pTHX) { return goperl_pp_dispatch(aTHX_ 140); }
static OP *goperl_pp_slot_141(pTHX) { return goperl_pp_dispatch(aTHX_ 141); }
static OP *goperl_pp_slot_142(pTHX) { return goperl_pp_dispatch(aTHX_ 142); }
static OP *goperl_pp_slot_143(pTHX) { return goperl_pp_dispatch(aTHX_ 143); }
static OP *goperl_pp_slot_144(pTHX) { return goperl_pp_dispatch(aTHX_ 144); }
static OP *goperl_pp_slot_145(pTHX) { return goperl_pp_dispatch(aTHX_ 145); }
static OP *goperl_pp_slot_146(pTHX) { return goperl_pp_dispatch(aTHX_ 146); }
static OP *goperl_pp_slot_147(pTHX) { return goperl_pp_dispatch(aTHX_ 147); }
static OP *goperl_pp_slot_148(pTHX) { return goperl_pp_dispatch(aTHX_ 148); }
static OP *goperl_pp_slot_149(pTHX) { return goperl_pp_dispatch(aTHX_ 149); }
static OP *goperl_pp_slot_150(pTHX) { return goperl_pp_dispatch(aTHX_ 150); }
static OP *goperl_pp_slot_151(pTHX) { return goperl_pp_dispatch(aTHX_ 151); }
static OP *goperl_pp_slot_152(pTHX) { return goperl_pp_dispatch(aTHX_ 152); }
static OP *goperl_pp_slot_153(pTHX) { return goperl_pp_dispatch(aTHX_ 153); }
static OP *goperl_pp_slot_154(pTHX) { return goperl_pp_dispatch(aTHX_ 154); }
static OP *goperl_pp_slot_155(pTHX) { return goperl_pp_dispatch(aTHX_ 155); }
static OP *goperl_pp_slot_156(pTHX) { return goperl_pp_dispatch(aTHX_ 156); }
static OP *goperl_pp_slot_157(pTHX) { return goperl_pp_dispatch(aTHX_ 157); }
static OP *goperl_pp_slot_158(pTHX) { return goperl_pp_dispatch(aTHX_ 158); }
static OP *goperl_pp_slot_159(pTHX) { return goperl_pp_dispatch(aTHX_ 159); }
static OP *goperl_pp_slot_160(pTHX) { return goperl_pp_dispatch(aTHX_ 160); }
static OP *goperl_pp_slot_161(pTHX) { return goperl_pp_dispatch(aTHX_ 161); }
static OP *goperl_pp_slot_162(pTHX) { return goperl_pp_dispatch(aTHX_ 162); }
static OP *goperl_pp_slot_163(pTHX) { return goperl_pp_dispatch(aTHX_ 163); }
static OP *goperl_pp_slot_164(pTHX) { return goperl_pp_dispatch(aTHX_ 164); }
static OP *goperl_pp_slot_165(pTHX) { return goperl_pp_dispatch(aTHX_ 165); }
static OP *goperl_pp_slot_166(pTHX) { return goperl_pp_dispatch(aTHX_ 166); }
static OP *goperl_pp_slot_167(pTHX) { return goperl_pp_dispatch(aTHX_ 167); }
static OP *goperl_pp_slot_168(pTHX) { return goperl_pp_dispatch(aTHX_ 168); }
static OP *goperl_pp_slot_169(pTHX) { return goperl_pp_dispatch(aTHX_ 169); }
static OP *goperl_pp_slot_170(pTHX) { return goperl_pp_dispatch(aTHX_ 170); }
static OP *goperl_pp_slot_171(pTHX) { return goperl_pp_dispatch(aTHX_ 171); }
static OP *goperl_pp_slot_172(pTHX) { return goperl_pp_dispatch(aTHX_ 172); }
static OP *goperl_pp_slot_173(pTHX) { return goperl_pp_dispatch(aTHX_ 173); }
static OP *goperl_pp_slot_174(pTHX) { return goperl_pp_dispatch(aTHX_ 174); }
static OP *goperl_pp_slot_175(pTHX) { return goperl_pp_dispatch(aTHX_ 175); }
static OP *goperl_pp_slot_176(pTHX) { return goperl_pp_dispatch(aTHX_ 176); }
static OP *goperl_pp_slot_177(pTHX) { return goperl_pp_dispatch(aTHX_ 177); }
static OP *goperl_pp_slot_178(pTHX) { return goperl_pp_dispatch(aTHX_ 178); }
static OP *goperl_pp_slot_179(pTHX) { return goperl_pp_dispatch(aTHX_ 179); }
static OP *goperl_pp_slot_180(pTHX) { return goperl_pp_dispatch(aTHX_ 180); }
static OP *goperl_pp_slot_181(pTHX) { return goperl_pp_dispatch(aTHX_ 181); }
static OP *goperl_pp_slot_182(pTHX) { return goperl_pp_dispatch(aTHX_ 182); }
static OP *goperl_pp_slot_183(pTHX) { return goperl_pp_dispatch(aTHX_ 183); }
static OP *goperl_pp_slot_184(pTHX) { return goperl_pp_dispatch(aTHX_ 184); }
static OP *goperl_pp_slot_185(pTHX) { return goperl_pp_dispatch(aTHX_ 185); }
static OP *goperl_pp_slot_186(pTHX) { return goperl_pp_dispatch(aTHX_ 186); }
static OP *goperl_pp_slot_187(pTHX) { return goperl_pp_dispatch(aTHX_ 187); }
static OP *goperl_pp_slot_188(pTHX) { return goperl_pp_dispatch(aTHX_ 188); }
static OP *goperl_pp_slot_189(pTHX) { return goperl_pp_dispatch(aTHX_ 189); }
static OP *goperl_pp_slot_190(pTHX) { return goperl_pp_dispatch(aTHX_ 190); }
static OP *goperl_pp_slot_191(pTHX) { return goperl_pp_dispatch(aTHX_ 191); }
static OP *goperl_pp_slot_192(pTHX) { return goperl_pp_dispatch(aTHX_ 192); }
static OP *goperl_pp_slot_193(pTHX) { return goperl_pp_dispatch(aTHX_ 193); }
static OP *goperl_pp_slot_194(pTHX) { return goperl_pp_dispatch(aTHX_ 194); }
static OP *goperl_pp_slot_195(pTHX) { return goperl_pp_dispatch(aTHX_ 195); }
static OP *goperl_pp_slot_196(pTHX) { return goperl_pp_dispatch(aTHX_ 196); }
static OP *goperl_pp_slot_197(pTHX) { return goperl_pp_dispatch(aTHX_ 197); }
static OP *goperl_pp_slot_198(pTHX) { return goperl_pp_dispatch(aTHX_ 198); }
static OP *goperl_pp_slot_199(pTHX) { return goperl_pp_dispatch(aTHX_ 199); }
static OP *goperl_pp_slot_200(pTHX) { return goperl_pp_dispatch(aTHX_ 200); }
static OP *goperl_pp_slot_201(pTHX) { return goperl_pp_dispatch(aTHX_ 201); }
static OP *goperl_pp_slot_202(pTHX) { return goperl_pp_dispatch(aTHX_ 202); }
static OP *goperl_pp_slot_203(pTHX) { return goperl_pp_dispatch(aTHX_ 203); }
static OP *goperl_pp_slot_204(pTHX) { return goperl_pp_dispatch(aTHX_ 204); }
static OP *goperl_pp_slot_205(pTHX) { return goperl_pp_dispatch(aTHX_ 205); }
static OP *goperl_pp_slot_206(pTHX) { return goperl_pp_dispatch(aTHX_ 206); }
static OP *goperl_pp_slot_207(pTHX) { return goperl_pp_dispatch(aTHX_ 207); }
static OP *goperl_pp_slot_208(pTHX) { return goperl_pp_dispatch(aTHX_ 208); }
static OP *goperl_pp_slot_209(pTHX) { return goperl_pp_dispatch(aTHX_ 209); }
static OP *goperl_pp_slot_210(pTHX) { return goperl_pp_dispatch(aTHX_ 210); }
static OP *goperl_pp_slot_211(pTHX) { return goperl_pp_dispatch(aTHX_ 211); }
static OP *goperl_pp_slot_212(pTHX) { return goperl_pp_dispatch(aTHX_ 212); }
static OP *goperl_pp_slot_213(pTHX) { return goperl_pp_dispatch(aTHX_ 213); }
static OP *goperl_pp_slot_214(pTHX) { return goperl_pp_dispatch(aTHX_ 214); }
static OP *goperl_pp_slot_215(pTHX) { return goperl_pp_dispatch(aTHX_ 215); }
static OP *goperl_pp_slot_216(pTHX) { return goperl_pp_dispatch(aTHX_ 216); }
static OP *goperl_pp_slot_217(pTHX) { return goperl_pp_dispatch(aTHX_ 217); }
static OP *goperl_pp_slot_218(pTHX) { return goperl_pp_dispatch(aTHX_ 218); }
static OP *goperl_pp_slot_219(pTHX) { return goperl_pp_dispatch(aTHX_ 219); }
static OP *goperl_pp_slot_220(pTHX) { return goperl_pp_dispatch(aTHX_ 220); }
static OP *goperl_pp_slot_221(pTHX) { return goperl_pp_dispatch(aTHX_ 221); }
static OP *goperl_pp_slot_222(pTHX) { return goperl_pp_dispatch(aTHX_ 222); }
static OP *goperl_pp_slot_223(pTHX) { return goperl_pp_dispatch(aTHX_ 223); }
static OP *goperl_pp_slot_224(pTHX) { return goperl_pp_dispatch(aTHX_ 224); }
static OP *goperl_pp_slot_225(pTHX) { return goperl_pp_dispatch(aTHX_ 225); }
static OP *goperl_pp_slot_226(pTHX) { return goperl_pp_dispatch(aTHX_ 226); }
static OP *goperl_pp_slot_227(pTHX) { return goperl_pp_dispatch(aTHX_ 227); }
static OP *goperl_pp_slot_228(pTHX) { return goperl_pp_dispatch(aTHX_ 228); }
static OP *goperl_pp_slot_229(pTHX) { return goperl_pp_dispatch(aTHX_ 229); }
static OP *goperl_pp_slot_230(pTHX) { return goperl_pp_dispatch(aTHX_ 230); }
static OP *goperl_pp_slot_231(pTHX) { return goperl_pp_dispatch(aTHX_ 231); }
static OP *goperl_pp_slot_232(pTHX) { return goperl_pp_dispatch(aTHX_ 232); }
static OP *goperl_pp_slot_233(pTHX) { return goperl_pp_dispatch(aTHX_ 233); }
static OP *goperl_pp_slot_234(pTHX) { return goperl_pp_dispatch(aTHX_ 234); }
static OP *goperl_pp_slot_235(pTHX) { return goperl_pp_dispatch(aTHX_ 235); }
static OP *goperl_pp_slot_236(pTHX) { return goperl_pp_dispatch(aTHX_ 236); }
static OP *goperl_pp_slot_237(pTHX) { return goperl_pp_dispatch(aTHX_ 237); }
static OP *goperl_pp_slot_238(pTHX) { return goperl_pp_dispatch(aTHX_ 238); }
static OP *goperl_pp_slot_239(pTHX) { return goperl_pp_dispatch(aTHX_ 239); }
static OP *goperl_pp_slot_240(pTHX) { return goperl_pp_dispatch(aTHX_ 240); }
static OP *goperl_pp_slot_241(pTHX) { return goperl_pp_dispatch(aTHX_ 241); }
static OP *goperl_pp_slot_242(pTHX) { return goperl_pp_dispatch(aTHX_ 242); }
static OP *goperl_pp_slot_243(pTHX) { return goperl_pp_dispatch(aTHX_ 243); }
static OP *goperl_pp_slot_244(pTHX) { return goperl_pp_dispatch(aTHX_ 244); }
static OP *goperl_pp_slot_245(pTHX) { return goperl_pp_dispatch(aTHX_ 245); }
static OP *goperl_pp_slot_246(pTHX) { return goperl_pp_dispatch(aTHX_ 246); }
static OP *goperl_pp_slot_247(pTHX) { return goperl_pp_dispatch(aTHX_ 247); }
static OP *goperl_pp_slot_248(pTHX) { return goperl_pp_dispatch(aTHX_ 248); }
static OP *goperl_pp_slot_249(pTHX) { return goperl_pp_dispatch(aTHX_ 249); }
static OP *goperl_pp_slot_250(pTHX) { return goperl_pp_dispatch(aTHX_ 250); }
static OP *goperl_pp_slot_251(pTHX) { return goperl_pp_dispatch(aTHX_ 251); }
static OP *goperl_pp_slot_252(pTHX) { return goperl_pp_dispatch(aTHX_ 252); }
static OP *goperl_pp_slot_253(pTHX) { return goperl_pp_dispatch(aTHX_ 253); }
static OP *goperl_pp_slot_254(pTHX) { return goperl_pp_dispatch(aTHX_ 254); }
static OP *goperl_pp_slot_255(pTHX) { return goperl_pp_dispatch(aTHX_ 255); }
static OP *goperl_pp_slot_256(pTHX) { return goperl_pp_dispatch(aTHX_ 256); }
static OP *goperl_pp_slot_257(pTHX) { return goperl_pp_dispatch(aTHX_ 257); }
static OP *goperl_pp_slot_258(pTHX) { return goperl_pp_dispatch(aTHX_ 258); }
static OP *goperl_pp_slot_259(pTHX) { return goperl_pp_dispatch(aTHX_ 259); }
static OP *goperl_pp_slot_260(pTHX) { return goperl_pp_dispatch(aTHX_ 260); }
static OP *goperl_pp_slot_261(pTHX) { return goperl_pp_dispatch(aTHX_ 261); }
static OP *goperl_pp_slot_262(pTHX) { return goperl_pp_dispatch(aTHX_ 262); }
static OP *goperl_pp_slot_263(pTHX) { return goperl_pp_dispatch(aTHX_ 263); }
static OP *goperl_pp_slot_264(pTHX) { return goperl_pp_dispatch(aTHX_ 264); }
static OP *goperl_pp_slot_265(pTHX) { return goperl_pp_dispatch(aTHX_ 265); }
static OP *goperl_pp_slot_266(pTHX) { return goperl_pp_dispatch(aTHX_ 266); }
static OP *goperl_pp_slot_267(pTHX) { return goperl_pp_dispatch(aTHX_ 267); }
static OP *goperl_pp_slot_268(pTHX) { return goperl_pp_dispatch(aTHX_ 268); }
static OP *goperl_pp_slot_269(pTHX) { return goperl_pp_dispatch(aTHX_ 269); }
static OP *goperl_pp_slot_270(pTHX) { return goperl_pp_dispatch(aTHX_ 270); }
static OP *goperl_pp_slot_271(pTHX) { return goperl_pp_dispatch(aTHX_ 271); }
static OP *goperl_pp_slot_272(pTHX) { return goperl_pp_dispatch(aTHX_ 272); }
static OP *goperl_pp_slot_273(pTHX) { return goperl_pp_dispatch(aTHX_ 273); }
static OP *goperl_pp_slot_274(pTHX) { return goperl_pp_dispatch(aTHX_ 274); }
static OP *goperl_pp_slot_275(pTHX) { return goperl_pp_dispatch(aTHX_ 275); }
static OP *goperl_pp_slot_276(pTHX) { return goperl_pp_dispatch(aTHX_ 276); }
static OP *goperl_pp_slot_277(pTHX) { return goperl_pp_dispatch(aTHX_ 277); }
static OP *goperl_pp_slot_278(pTHX) { return goperl_pp_dispatch(aTHX_ 278); }
static OP *goperl_pp_slot_279(pTHX) { return goperl_pp_dispatch(aTHX_ 279); }
static OP *goperl_pp_slot_280(pTHX) { return goperl_pp_dispatch(aTHX_ 280); }
static OP *goperl_pp_slot_281(pTHX) { return goperl_pp_dispatch(aTHX_ 281); }
static OP *goperl_pp_slot_282(pTHX) { return goperl_pp_dispatch(aTHX_ 282); }
static OP *goperl_pp_slot_283(pTHX) { return goperl_pp_dispatch(aTHX_ 283); }
static OP *goperl_pp_slot_284(pTHX) { return goperl_pp_dispatch(aTHX_ 284); }
static OP *goperl_pp_slot_285(pTHX) { return goperl_pp_dispatch(aTHX_ 285); }
static OP *goperl_pp_slot_286(pTHX) { return goperl_pp_dispatch(aTHX_ 286); }
static OP *goperl_pp_slot_287(pTHX) { return goperl_pp_dispatch(aTHX_ 287); }
static OP *goperl_pp_slot_288(pTHX) { return goperl_pp_dispatch(aTHX_ 288); }
static OP *goperl_pp_slot_289(pTHX) { return goperl_pp_dispatch(aTHX_ 289); }
static OP *goperl_pp_slot_290(pTHX) { return goperl_pp_dispatch(aTHX_ 290); }
static OP *goperl_pp_slot_291(pTHX) { return goperl_pp_dispatch(aTHX_ 291); }
static OP *goperl_pp_slot_292(pTHX) { return goperl_pp_dispatch(aTHX_ 292); }
static OP *goperl_pp_slot_293(pTHX) { return goperl_pp_dispatch(aTHX_ 293); }
static OP *goperl_pp_slot_294(pTHX) { return goperl_pp_dispatch(aTHX_ 294); }
static OP *goperl_pp_slot_295(pTHX) { return goperl_pp_dispatch(aTHX_ 295); }
static OP *goperl_pp_slot_296(pTHX) { return goperl_pp_dispatch(aTHX_ 296); }
static OP *goperl_pp_slot_297(pTHX) { return goperl_pp_dispatch(aTHX_ 297); }
static OP *goperl_pp_slot_298(pTHX) { return goperl_pp_dispatch(aTHX_ 298); }
static OP *goperl_pp_slot_299(pTHX) { return goperl_pp_dispatch(aTHX_ 299); }
static OP *goperl_pp_slot_300(pTHX) { return goperl_pp_dispatch(aTHX_ 300); }
static OP *goperl_pp_slot_301(pTHX) { return goperl_pp_dispatch(aTHX_ 301); }
static OP *goperl_pp_slot_302(pTHX) { return goperl_pp_dispatch(aTHX_ 302); }
static OP *goperl_pp_slot_303(pTHX) { return goperl_pp_dispatch(aTHX_ 303); }
static OP *goperl_pp_slot_304(pTHX) { return goperl_pp_dispatch(aTHX_ 304); }
static OP *goperl_pp_slot_305(pTHX) { return goperl_pp_dispatch(aTHX_ 305); }
static OP *goperl_pp_slot_306(pTHX) { return goperl_pp_dispatch(aTHX_ 306); }
static OP *goperl_pp_slot_307(pTHX) { return goperl_pp_dispatch(aTHX_ 307); }
static OP *goperl_pp_slot_308(pTHX) { return goperl_pp_dispatch(aTHX_ 308); }
static OP *goperl_pp_slot_309(pTHX) { return goperl_pp_dispatch(aTHX_ 309); }
static OP *goperl_pp_slot_310(pTHX) { return goperl_pp_dispatch(aTHX_ 310); }
static OP *goperl_pp_slot_311(pTHX) { return goperl_pp_dispatch(aTHX_ 311); }
static OP *goperl_pp_slot_312(pTHX) { return goperl_pp_dispatch(aTHX_ 312); }
static OP *goperl_pp_slot_313(pTHX) { return goperl_pp_dispatch(aTHX_ 313); }
static OP *goperl_pp_slot_314(pTHX) { return goperl_pp_dispatch(aTHX_ 314); }
static OP *goperl_pp_slot_315(pTHX) { return goperl_pp_dispatch(aTHX_ 315); }
static OP *goperl_pp_slot_316(pTHX) { return goperl_pp_dispatch(aTHX_ 316); }
static OP *goperl_pp_slot_317(pTHX) { return goperl_pp_dispatch(aTHX_ 317); }
static OP *goperl_pp_slot_318(pTHX) { return goperl_pp_dispatch(aTHX_ 318); }
static OP *goperl_pp_slot_319(pTHX) { return goperl_pp_dispatch(aTHX_ 319); }
static OP *goperl_pp_slot_320(pTHX) { return goperl_pp_dispatch(aTHX_ 320); }
static OP *goperl_pp_slot_321(pTHX) { return goperl_pp_dispatch(aTHX_ 321); }
static OP *goperl_pp_slot_322(pTHX) { return goperl_pp_dispatch(aTHX_ 322); }
static OP *goperl_pp_slot_323(pTHX) { return goperl_pp_dispatch(aTHX_ 323); }
static OP *goperl_pp_slot_324(pTHX) { return goperl_pp_dispatch(aTHX_ 324); }
static OP *goperl_pp_slot_325(pTHX) { return goperl_pp_dispatch(aTHX_ 325); }
static OP *goperl_pp_slot_326(pTHX) { return goperl_pp_dispatch(aTHX_ 326); }
static OP *goperl_pp_slot_327(pTHX) { return goperl_pp_dispatch(aTHX_ 327); }
static OP *goperl_pp_slot_328(pTHX) { return goperl_pp_dispatch(aTHX_ 328); }
static OP *goperl_pp_slot_329(pTHX) { return goperl_pp_dispatch(aTHX_ 329); }
static OP *goperl_pp_slot_330(pTHX) { return goperl_pp_dispatch(aTHX_ 330); }
static OP *goperl_pp_slot_331(pTHX) { return goperl_pp_dispatch(aTHX_ 331); }
static OP *goperl_pp_slot_332(pTHX) { return goperl_pp_dispatch(aTHX_ 332); }
static OP *goperl_pp_slot_333(pTHX) { return goperl_pp_dispatch(aTHX_ 333); }
static OP *goperl_pp_slot_334(pTHX) { return goperl_pp_dispatch(aTHX_ 334); }
static OP *goperl_pp_slot_335(pTHX) { return goperl_pp_dispatch(aTHX_ 335); }
static OP *goperl_pp_slot_336(pTHX) { return goperl_pp_dispatch(aTHX_ 336); }
static OP *goperl_pp_slot_337(pTHX) { return goperl_pp_dispatch(aTHX_ 337); }
static OP *goperl_pp_slot_338(pTHX) { return goperl_pp_dispatch(aTHX_ 338); }
static OP *goperl_pp_slot_339(pTHX) { return goperl_pp_dispatch(aTHX_ 339); }
static OP *goperl_pp_slot_340(pTHX) { return goperl_pp_dispatch(aTHX_ 340); }
static OP *goperl_pp_slot_341(pTHX) { return goperl_pp_dispatch(aTHX_ 341); }
static OP *goperl_pp_slot_342(pTHX) { return goperl_pp_dispatch(aTHX_ 342); }
static OP *goperl_pp_slot_343(pTHX) { return goperl_pp_dispatch(aTHX_ 343); }
static OP *goperl_pp_slot_344(pTHX) { return goperl_pp_dispatch(aTHX_ 344); }
static OP *goperl_pp_slot_345(pTHX) { return goperl_pp_dispatch(aTHX_ 345); }
static OP *goperl_pp_slot_346(pTHX) { return goperl_pp_dispatch(aTHX_ 346); }
static OP *goperl_pp_slot_347(pTHX) { return goperl_pp_dispatch(aTHX_ 347); }
static OP *goperl_pp_slot_348(pTHX) { return goperl_pp_dispatch(aTHX_ 348); }
static OP *goperl_pp_slot_349(pTHX) { return goperl_pp_dispatch(aTHX_ 349); }
static OP *goperl_pp_slot_350(pTHX) { return goperl_pp_dispatch(aTHX_ 350); }
static OP *goperl_pp_slot_351(pTHX) { return goperl_pp_dispatch(aTHX_ 351); }
static OP *goperl_pp_slot_352(pTHX) { return goperl_pp_dispatch(aTHX_ 352); }
static OP *goperl_pp_slot_353(pTHX) { return goperl_pp_dispatch(aTHX_ 353); }
static OP *goperl_pp_slot_354(pTHX) { return goperl_pp_dispatch(aTHX_ 354); }
static OP *goperl_pp_slot_355(pTHX) { return goperl_pp_dispatch(aTHX_ 355); }
static OP *goperl_pp_slot_356(pTHX) { return goperl_pp_dispatch(aTHX_ 356); }
static OP *goperl_pp_slot_357(pTHX) { return goperl_pp_dispatch(aTHX_ 357); }
static OP *goperl_pp_slot_358(pTHX) { return goperl_pp_dispatch(aTHX_ 358); }
static OP *goperl_pp_slot_359(pTHX) { return goperl_pp_dispatch(aTHX_ 359); }
static OP *goperl_pp_slot_360(pTHX) { return goperl_pp_dispatch(aTHX_ 360); }
static OP *goperl_pp_slot_361(pTHX) { return goperl_pp_dispatch(aTHX_ 361); }
static OP *goperl_pp_slot_362(pTHX) { return goperl_pp_dispatch(aTHX_ 362); }
static OP *goperl_pp_slot_363(pTHX) { return goperl_pp_dispatch(aTHX_ 363); }
static OP *goperl_pp_slot_364(pTHX) { return goperl_pp_dispatch(aTHX_ 364); }
static OP *goperl_pp_slot_365(pTHX) { return goperl_pp_dispatch(aTHX_ 365); }
static OP *goperl_pp_slot_366(pTHX) { return goperl_pp_dispatch(aTHX_ 366); }
static OP *goperl_pp_slot_367(pTHX) { return goperl_pp_dispatch(aTHX_ 367); }
static OP *goperl_pp_slot_368(pTHX) { return goperl_pp_dispatch(aTHX_ 368); }
static OP *goperl_pp_slot_369(pTHX) { return goperl_pp_dispatch(aTHX_ 369); }
static OP *goperl_pp_slot_370(pTHX) { return goperl_pp_dispatch(aTHX_ 370); }
static OP *goperl_pp_slot_371(pTHX) { return goperl_pp_dispatch(aTHX_ 371); }
static OP *goperl_pp_slot_372(pTHX) { return goperl_pp_dispatch(aTHX_ 372); }
static OP *goperl_pp_slot_373(pTHX) { return goperl_pp_dispatch(aTHX_ 373); }
static OP *goperl_pp_slot_374(pTHX) { return goperl_pp_dispatch(aTHX_ 374); }
static OP *goperl_pp_slot_375(pTHX) { return goperl_pp_dispatch(aTHX_ 375); }
static OP *goperl_pp_slot_376(pTHX) { return goperl_pp_dispatch(aTHX_ 376); }
static OP *goperl_pp_slot_377(pTHX) { return goperl_pp_dispatch(aTHX_ 377); }
static OP *goperl_pp_slot_378(pTHX) { return goperl_pp_dispatch(aTHX_ 378); }
static OP *goperl_pp_slot_379(pTHX) { return goperl_pp_dispatch(aTHX_ 379); }
static OP *goperl_pp_slot_380(pTHX) { return goperl_pp_dispatch(aTHX_ 380); }
static OP *goperl_pp_slot_381(pTHX) { return goperl_pp_dispatch(aTHX_ 381); }
static OP *goperl_pp_slot_382(pTHX) { return goperl_pp_dispatch(aTHX_ 382); }
static OP *goperl_pp_slot_383(pTHX) { return goperl_pp_dispatch(aTHX_ 383); }
static OP *goperl_pp_slot_384(pTHX) { return goperl_pp_dispatch(aTHX_ 384); }
static OP *goperl_pp_slot_385(pTHX) { return goperl_pp_dispatch(aTHX_ 385); }
static OP *goperl_pp_slot_386(pTHX) { return goperl_pp_dispatch(aTHX_ 386); }
static OP *goperl_pp_slot_387(pTHX) { return goperl_pp_dispatch(aTHX_ 387); }
static OP *goperl_pp_slot_388(pTHX) { return goperl_pp_dispatch(aTHX_ 388); }
static OP *goperl_pp_slot_389(pTHX) { return goperl_pp_dispatch(aTHX_ 389); }
static OP *goperl_pp_slot_390(pTHX) { return goperl_pp_dispatch(aTHX_ 390); }
static OP *goperl_pp_slot_391(pTHX) { return goperl_pp_dispatch(aTHX_ 391); }
static OP *goperl_pp_slot_392(pTHX) { return goperl_pp_dispatch(aTHX_ 392); }
static OP *goperl_pp_slot_393(pTHX) { return goperl_pp_dispatch(aTHX_ 393); }
static OP *goperl_pp_slot_394(pTHX) { return goperl_pp_dispatch(aTHX_ 394); }
static OP *goperl_pp_slot_395(pTHX) { return goperl_pp_dispatch(aTHX_ 395); }
static OP *goperl_pp_slot_396(pTHX) { return goperl_pp_dispatch(aTHX_ 396); }
static OP *goperl_pp_slot_397(pTHX) { return goperl_pp_dispatch(aTHX_ 397); }
static OP *goperl_pp_slot_398(pTHX) { return goperl_pp_dispatch(aTHX_ 398); }
static OP *goperl_pp_slot_399(pTHX) { return goperl_pp_dispatch(aTHX_ 399); }
static OP *goperl_pp_slot_400(pTHX) { return goperl_pp_dispatch(aTHX_ 400); }
static OP *goperl_pp_slot_401(pTHX) { return goperl_pp_dispatch(aTHX_ 401); }
static OP *goperl_pp_slot_402(pTHX) { return goperl_pp_dispatch(aTHX_ 402); }
static OP *goperl_pp_slot_403(pTHX) { return goperl_pp_dispatch(aTHX_ 403); }
static OP *goperl_pp_slot_404(pTHX) { return goperl_pp_dispatch(aTHX_ 404); }
static OP *goperl_pp_slot_405(pTHX) { return goperl_pp_dispatch(aTHX_ 405); }
static OP *goperl_pp_slot_406(pTHX) { return goperl_pp_dispatch(aTHX_ 406); }
static OP *goperl_pp_slot_407(pTHX) { return goperl_pp_dispatch(aTHX_ 407); }
static OP *goperl_pp_slot_408(pTHX) { return goperl_pp_dispatch(aTHX_ 408); }
static OP *goperl_pp_slot_409(pTHX) { return goperl_pp_dispatch(aTHX_ 409); }
static OP *goperl_pp_slot_410(pTHX) { return goperl_pp_dispatch(aTHX_ 410); }
static OP *goperl_pp_slot_411(pTHX) { return goperl_pp_dispatch(aTHX_ 411); }
static OP *goperl_pp_slot_412(pTHX) { return goperl_pp_dispatch(aTHX_ 412); }
static OP *goperl_pp_slot_413(pTHX) { return goperl_pp_dispatch(aTHX_ 413); }
static OP *goperl_pp_slot_414(pTHX) { return goperl_pp_dispatch(aTHX_ 414); }
static OP *goperl_pp_slot_415(pTHX) { return goperl_pp_dispatch(aTHX_ 415); }
static OP *goperl_pp_slot_416(pTHX) { return goperl_pp_dispatch(aTHX_ 416); }
static OP *goperl_pp_slot_417(pTHX) { return goperl_pp_dispatch(aTHX_ 417); }
static OP *goperl_pp_slot_418(pTHX) { return goperl_pp_dispatch(aTHX_ 418); }
static OP *goperl_pp_slot_419(pTHX) { return goperl_pp_dispatch(aTHX_ 419); }
static OP *goperl_pp_slot_420(pTHX) { return goperl_pp_dispatch(aTHX_ 420); }
static OP *goperl_pp_slot_421(pTHX) { return goperl_pp_dispatch(aTHX_ 421); }
static OP *goperl_pp_slot_422(pTHX) { return goperl_pp_dispatch(aTHX_ 422); }
static OP *goperl_pp_slot_423(pTHX) { return goperl_pp_dispatch(aTHX_ 423); }
static OP *goperl_pp_slot_424(pTHX) { return goperl_pp_dispatch(aTHX_ 424); }
static OP *goperl_pp_slot_425(pTHX) { return goperl_pp_dispatch(aTHX_ 425); }

#define GOPERL_PP_SLOT_INIT(tbl) do { \
    (tbl)[0] = goperl_pp_slot_0; \
    (tbl)[1] = goperl_pp_slot_1; \
    (tbl)[2] = goperl_pp_slot_2; \
    (tbl)[3] = goperl_pp_slot_3; \
    (tbl)[4] = goperl_pp_slot_4; \
    (tbl)[5] = goperl_pp_slot_5; \
    (tbl)[6] = goperl_pp_slot_6; \
    (tbl)[7] = goperl_pp_slot_7; \
    (tbl)[8] = goperl_pp_slot_8; \
    (tbl)[9] = goperl_pp_slot_9; \
    (tbl)[10] = goperl_pp_slot_10; \
    (tbl)[11] = goperl_pp_slot_11; \
    (tbl)[12] = goperl_pp_slot_12; \
    (tbl)[13] = goperl_pp_slot_13; \
    (tbl)[14] = goperl_pp_slot_14; \
    (tbl)[15] = goperl_pp_slot_15; \
    (tbl)[16] = goperl_pp_slot_16; \
    (tbl)[17] = goperl_pp_slot_17; \
    (tbl)[18] = goperl_pp_slot_18; \
    (tbl)[19] = goperl_pp_slot_19; \
    (tbl)[20] = goperl_pp_slot_20; \
    (tbl)[21] = goperl_pp_slot_21; \
    (tbl)[22] = goperl_pp_slot_22; \
    (tbl)[23] = goperl_pp_slot_23; \
    (tbl)[24] = goperl_pp_slot_24; \
    (tbl)[25] = goperl_pp_slot_25; \
    (tbl)[26] = goperl_pp_slot_26; \
    (tbl)[27] = goperl_pp_slot_27; \
    (tbl)[28] = goperl_pp_slot_28; \
    (tbl)[29] = goperl_pp_slot_29; \
    (tbl)[30] = goperl_pp_slot_30; \
    (tbl)[31] = goperl_pp_slot_31; \
    (tbl)[32] = goperl_pp_slot_32; \
    (tbl)[33] = goperl_pp_slot_33; \
    (tbl)[34] = goperl_pp_slot_34; \
    (tbl)[35] = goperl_pp_slot_35; \
    (tbl)[36] = goperl_pp_slot_36; \
    (tbl)[37] = goperl_pp_slot_37; \
    (tbl)[38] = goperl_pp_slot_38; \
    (tbl)[39] = goperl_pp_slot_39; \
    (tbl)[40] = goperl_pp_slot_40; \
    (tbl)[41] = goperl_pp_slot_41; \
    (tbl)[42] = goperl_pp_slot_42; \
    (tbl)[43] = goperl_pp_slot_43; \
    (tbl)[44] = goperl_pp_slot_44; \
    (tbl)[45] = goperl_pp_slot_45; \
    (tbl)[46] = goperl_pp_slot_46; \
    (tbl)[47] = goperl_pp_slot_47; \
    (tbl)[48] = goperl_pp_slot_48; \
    (tbl)[49] = goperl_pp_slot_49; \
    (tbl)[50] = goperl_pp_slot_50; \
    (tbl)[51] = goperl_pp_slot_51; \
    (tbl)[52] = goperl_pp_slot_52; \
    (tbl)[53] = goperl_pp_slot_53; \
    (tbl)[54] = goperl_pp_slot_54; \
    (tbl)[55] = goperl_pp_slot_55; \
    (tbl)[56] = goperl_pp_slot_56; \
    (tbl)[57] = goperl_pp_slot_57; \
    (tbl)[58] = goperl_pp_slot_58; \
    (tbl)[59] = goperl_pp_slot_59; \
    (tbl)[60] = goperl_pp_slot_60; \
    (tbl)[61] = goperl_pp_slot_61; \
    (tbl)[62] = goperl_pp_slot_62; \
    (tbl)[63] = goperl_pp_slot_63; \
    (tbl)[64] = goperl_pp_slot_64; \
    (tbl)[65] = goperl_pp_slot_65; \
    (tbl)[66] = goperl_pp_slot_66; \
    (tbl)[67] = goperl_pp_slot_67; \
    (tbl)[68] = goperl_pp_slot_68; \
    (tbl)[69] = goperl_pp_slot_69; \
    (tbl)[70] = goperl_pp_slot_70; \
    (tbl)[71] = goperl_pp_slot_71; \
    (tbl)[72] = goperl_pp_slot_72; \
    (tbl)[73] = goperl_pp_slot_73; \
    (tbl)[74] = goperl_pp_slot_74; \
    (tbl)[75] = goperl_pp_slot_75; \
    (tbl)[76] = goperl_pp_slot_76; \
    (tbl)[77] = goperl_pp_slot_77; \
    (tbl)[78] = goperl_pp_slot_78; \
    (tbl)[79] = goperl_pp_slot_79; \
    (tbl)[80] = goperl_pp_slot_80; \
    (tbl)[81] = goperl_pp_slot_81; \
    (tbl)[82] = goperl_pp_slot_82; \
    (tbl)[83] = goperl_pp_slot_83; \
    (tbl)[84] = goperl_pp_slot_84; \
    (tbl)[85] = goperl_pp_slot_85; \
    (tbl)[86] = goperl_pp_slot_86; \
    (tbl)[87] = goperl_pp_slot_87; \
    (tbl)[88] = goperl_pp_slot_88; \
    (tbl)[89] = goperl_pp_slot_89; \
    (tbl)[90] = goperl_pp_slot_90; \
    (tbl)[91] = goperl_pp_slot_91; \
    (tbl)[92] = goperl_pp_slot_92; \
    (tbl)[93] = goperl_pp_slot_93; \
    (tbl)[94] = goperl_pp_slot_94; \
    (tbl)[95] = goperl_pp_slot_95; \
    (tbl)[96] = goperl_pp_slot_96; \
    (tbl)[97] = goperl_pp_slot_97; \
    (tbl)[98] = goperl_pp_slot_98; \
    (tbl)[99] = goperl_pp_slot_99; \
    (tbl)[100] = goperl_pp_slot_100; \
    (tbl)[101] = goperl_pp_slot_101; \
    (tbl)[102] = goperl_pp_slot_102; \
    (tbl)[103] = goperl_pp_slot_103; \
    (tbl)[104] = goperl_pp_slot_104; \
    (tbl)[105] = goperl_pp_slot_105; \
    (tbl)[106] = goperl_pp_slot_106; \
    (tbl)[107] = goperl_pp_slot_107; \
    (tbl)[108] = goperl_pp_slot_108; \
    (tbl)[109] = goperl_pp_slot_109; \
    (tbl)[110] = goperl_pp_slot_110; \
    (tbl)[111] = goperl_pp_slot_111; \
    (tbl)[112] = goperl_pp_slot_112; \
    (tbl)[113] = goperl_pp_slot_113; \
    (tbl)[114] = goperl_pp_slot_114; \
    (tbl)[115] = goperl_pp_slot_115; \
    (tbl)[116] = goperl_pp_slot_116; \
    (tbl)[117] = goperl_pp_slot_117; \
    (tbl)[118] = goperl_pp_slot_118; \
    (tbl)[119] = goperl_pp_slot_119; \
    (tbl)[120] = goperl_pp_slot_120; \
    (tbl)[121] = goperl_pp_slot_121; \
    (tbl)[122] = goperl_pp_slot_122; \
    (tbl)[123] = goperl_pp_slot_123; \
    (tbl)[124] = goperl_pp_slot_124; \
    (tbl)[125] = goperl_pp_slot_125; \
    (tbl)[126] = goperl_pp_slot_126; \
    (tbl)[127] = goperl_pp_slot_127; \
    (tbl)[128] = goperl_pp_slot_128; \
    (tbl)[129] = goperl_pp_slot_129; \
    (tbl)[130] = goperl_pp_slot_130; \
    (tbl)[131] = goperl_pp_slot_131; \
    (tbl)[132] = goperl_pp_slot_132; \
    (tbl)[133] = goperl_pp_slot_133; \
    (tbl)[134] = goperl_pp_slot_134; \
    (tbl)[135] = goperl_pp_slot_135; \
    (tbl)[136] = goperl_pp_slot_136; \
    (tbl)[137] = goperl_pp_slot_137; \
    (tbl)[138] = goperl_pp_slot_138; \
    (tbl)[139] = goperl_pp_slot_139; \
    (tbl)[140] = goperl_pp_slot_140; \
    (tbl)[141] = goperl_pp_slot_141; \
    (tbl)[142] = goperl_pp_slot_142; \
    (tbl)[143] = goperl_pp_slot_143; \
    (tbl)[144] = goperl_pp_slot_144; \
    (tbl)[145] = goperl_pp_slot_145; \
    (tbl)[146] = goperl_pp_slot_146; \
    (tbl)[147] = goperl_pp_slot_147; \
    (tbl)[148] = goperl_pp_slot_148; \
    (tbl)[149] = goperl_pp_slot_149; \
    (tbl)[150] = goperl_pp_slot_150; \
    (tbl)[151] = goperl_pp_slot_151; \
    (tbl)[152] = goperl_pp_slot_152; \
    (tbl)[153] = goperl_pp_slot_153; \
    (tbl)[154] = goperl_pp_slot_154; \
    (tbl)[155] = goperl_pp_slot_155; \
    (tbl)[156] = goperl_pp_slot_156; \
    (tbl)[157] = goperl_pp_slot_157; \
    (tbl)[158] = goperl_pp_slot_158; \
    (tbl)[159] = goperl_pp_slot_159; \
    (tbl)[160] = goperl_pp_slot_160; \
    (tbl)[161] = goperl_pp_slot_161; \
    (tbl)[162] = goperl_pp_slot_162; \
    (tbl)[163] = goperl_pp_slot_163; \
    (tbl)[164] = goperl_pp_slot_164; \
    (tbl)[165] = goperl_pp_slot_165; \
    (tbl)[166] = goperl_pp_slot_166; \
    (tbl)[167] = goperl_pp_slot_167; \
    (tbl)[168] = goperl_pp_slot_168; \
    (tbl)[169] = goperl_pp_slot_169; \
    (tbl)[170] = goperl_pp_slot_170; \
    (tbl)[171] = goperl_pp_slot_171; \
    (tbl)[172] = goperl_pp_slot_172; \
    (tbl)[173] = goperl_pp_slot_173; \
    (tbl)[174] = goperl_pp_slot_174; \
    (tbl)[175] = goperl_pp_slot_175; \
    (tbl)[176] = goperl_pp_slot_176; \
    (tbl)[177] = goperl_pp_slot_177; \
    (tbl)[178] = goperl_pp_slot_178; \
    (tbl)[179] = goperl_pp_slot_179; \
    (tbl)[180] = goperl_pp_slot_180; \
    (tbl)[181] = goperl_pp_slot_181; \
    (tbl)[182] = goperl_pp_slot_182; \
    (tbl)[183] = goperl_pp_slot_183; \
    (tbl)[184] = goperl_pp_slot_184; \
    (tbl)[185] = goperl_pp_slot_185; \
    (tbl)[186] = goperl_pp_slot_186; \
    (tbl)[187] = goperl_pp_slot_187; \
    (tbl)[188] = goperl_pp_slot_188; \
    (tbl)[189] = goperl_pp_slot_189; \
    (tbl)[190] = goperl_pp_slot_190; \
    (tbl)[191] = goperl_pp_slot_191; \
    (tbl)[192] = goperl_pp_slot_192; \
    (tbl)[193] = goperl_pp_slot_193; \
    (tbl)[194] = goperl_pp_slot_194; \
    (tbl)[195] = goperl_pp_slot_195; \
    (tbl)[196] = goperl_pp_slot_196; \
    (tbl)[197] = goperl_pp_slot_197; \
    (tbl)[198] = goperl_pp_slot_198; \
    (tbl)[199] = goperl_pp_slot_199; \
    (tbl)[200] = goperl_pp_slot_200; \
    (tbl)[201] = goperl_pp_slot_201; \
    (tbl)[202] = goperl_pp_slot_202; \
    (tbl)[203] = goperl_pp_slot_203; \
    (tbl)[204] = goperl_pp_slot_204; \
    (tbl)[205] = goperl_pp_slot_205; \
    (tbl)[206] = goperl_pp_slot_206; \
    (tbl)[207] = goperl_pp_slot_207; \
    (tbl)[208] = goperl_pp_slot_208; \
    (tbl)[209] = goperl_pp_slot_209; \
    (tbl)[210] = goperl_pp_slot_210; \
    (tbl)[211] = goperl_pp_slot_211; \
    (tbl)[212] = goperl_pp_slot_212; \
    (tbl)[213] = goperl_pp_slot_213; \
    (tbl)[214] = goperl_pp_slot_214; \
    (tbl)[215] = goperl_pp_slot_215; \
    (tbl)[216] = goperl_pp_slot_216; \
    (tbl)[217] = goperl_pp_slot_217; \
    (tbl)[218] = goperl_pp_slot_218; \
    (tbl)[219] = goperl_pp_slot_219; \
    (tbl)[220] = goperl_pp_slot_220; \
    (tbl)[221] = goperl_pp_slot_221; \
    (tbl)[222] = goperl_pp_slot_222; \
    (tbl)[223] = goperl_pp_slot_223; \
    (tbl)[224] = goperl_pp_slot_224; \
    (tbl)[225] = goperl_pp_slot_225; \
    (tbl)[226] = goperl_pp_slot_226; \
    (tbl)[227] = goperl_pp_slot_227; \
    (tbl)[228] = goperl_pp_slot_228; \
    (tbl)[229] = goperl_pp_slot_229; \
    (tbl)[230] = goperl_pp_slot_230; \
    (tbl)[231] = goperl_pp_slot_231; \
    (tbl)[232] = goperl_pp_slot_232; \
    (tbl)[233] = goperl_pp_slot_233; \
    (tbl)[234] = goperl_pp_slot_234; \
    (tbl)[235] = goperl_pp_slot_235; \
    (tbl)[236] = goperl_pp_slot_236; \
    (tbl)[237] = goperl_pp_slot_237; \
    (tbl)[238] = goperl_pp_slot_238; \
    (tbl)[239] = goperl_pp_slot_239; \
    (tbl)[240] = goperl_pp_slot_240; \
    (tbl)[241] = goperl_pp_slot_241; \
    (tbl)[242] = goperl_pp_slot_242; \
    (tbl)[243] = goperl_pp_slot_243; \
    (tbl)[244] = goperl_pp_slot_244; \
    (tbl)[245] = goperl_pp_slot_245; \
    (tbl)[246] = goperl_pp_slot_246; \
    (tbl)[247] = goperl_pp_slot_247; \
    (tbl)[248] = goperl_pp_slot_248; \
    (tbl)[249] = goperl_pp_slot_249; \
    (tbl)[250] = goperl_pp_slot_250; \
    (tbl)[251] = goperl_pp_slot_251; \
    (tbl)[252] = goperl_pp_slot_252; \
    (tbl)[253] = goperl_pp_slot_253; \
    (tbl)[254] = goperl_pp_slot_254; \
    (tbl)[255] = goperl_pp_slot_255; \
    (tbl)[256] = goperl_pp_slot_256; \
    (tbl)[257] = goperl_pp_slot_257; \
    (tbl)[258] = goperl_pp_slot_258; \
    (tbl)[259] = goperl_pp_slot_259; \
    (tbl)[260] = goperl_pp_slot_260; \
    (tbl)[261] = goperl_pp_slot_261; \
    (tbl)[262] = goperl_pp_slot_262; \
    (tbl)[263] = goperl_pp_slot_263; \
    (tbl)[264] = goperl_pp_slot_264; \
    (tbl)[265] = goperl_pp_slot_265; \
    (tbl)[266] = goperl_pp_slot_266; \
    (tbl)[267] = goperl_pp_slot_267; \
    (tbl)[268] = goperl_pp_slot_268; \
    (tbl)[269] = goperl_pp_slot_269; \
    (tbl)[270] = goperl_pp_slot_270; \
    (tbl)[271] = goperl_pp_slot_271; \
    (tbl)[272] = goperl_pp_slot_272; \
    (tbl)[273] = goperl_pp_slot_273; \
    (tbl)[274] = goperl_pp_slot_274; \
    (tbl)[275] = goperl_pp_slot_275; \
    (tbl)[276] = goperl_pp_slot_276; \
    (tbl)[277] = goperl_pp_slot_277; \
    (tbl)[278] = goperl_pp_slot_278; \
    (tbl)[279] = goperl_pp_slot_279; \
    (tbl)[280] = goperl_pp_slot_280; \
    (tbl)[281] = goperl_pp_slot_281; \
    (tbl)[282] = goperl_pp_slot_282; \
    (tbl)[283] = goperl_pp_slot_283; \
    (tbl)[284] = goperl_pp_slot_284; \
    (tbl)[285] = goperl_pp_slot_285; \
    (tbl)[286] = goperl_pp_slot_286; \
    (tbl)[287] = goperl_pp_slot_287; \
    (tbl)[288] = goperl_pp_slot_288; \
    (tbl)[289] = goperl_pp_slot_289; \
    (tbl)[290] = goperl_pp_slot_290; \
    (tbl)[291] = goperl_pp_slot_291; \
    (tbl)[292] = goperl_pp_slot_292; \
    (tbl)[293] = goperl_pp_slot_293; \
    (tbl)[294] = goperl_pp_slot_294; \
    (tbl)[295] = goperl_pp_slot_295; \
    (tbl)[296] = goperl_pp_slot_296; \
    (tbl)[297] = goperl_pp_slot_297; \
    (tbl)[298] = goperl_pp_slot_298; \
    (tbl)[299] = goperl_pp_slot_299; \
    (tbl)[300] = goperl_pp_slot_300; \
    (tbl)[301] = goperl_pp_slot_301; \
    (tbl)[302] = goperl_pp_slot_302; \
    (tbl)[303] = goperl_pp_slot_303; \
    (tbl)[304] = goperl_pp_slot_304; \
    (tbl)[305] = goperl_pp_slot_305; \
    (tbl)[306] = goperl_pp_slot_306; \
    (tbl)[307] = goperl_pp_slot_307; \
    (tbl)[308] = goperl_pp_slot_308; \
    (tbl)[309] = goperl_pp_slot_309; \
    (tbl)[310] = goperl_pp_slot_310; \
    (tbl)[311] = goperl_pp_slot_311; \
    (tbl)[312] = goperl_pp_slot_312; \
    (tbl)[313] = goperl_pp_slot_313; \
    (tbl)[314] = goperl_pp_slot_314; \
    (tbl)[315] = goperl_pp_slot_315; \
    (tbl)[316] = goperl_pp_slot_316; \
    (tbl)[317] = goperl_pp_slot_317; \
    (tbl)[318] = goperl_pp_slot_318; \
    (tbl)[319] = goperl_pp_slot_319; \
    (tbl)[320] = goperl_pp_slot_320; \
    (tbl)[321] = goperl_pp_slot_321; \
    (tbl)[322] = goperl_pp_slot_322; \
    (tbl)[323] = goperl_pp_slot_323; \
    (tbl)[324] = goperl_pp_slot_324; \
    (tbl)[325] = goperl_pp_slot_325; \
    (tbl)[326] = goperl_pp_slot_326; \
    (tbl)[327] = goperl_pp_slot_327; \
    (tbl)[328] = goperl_pp_slot_328; \
    (tbl)[329] = goperl_pp_slot_329; \
    (tbl)[330] = goperl_pp_slot_330; \
    (tbl)[331] = goperl_pp_slot_331; \
    (tbl)[332] = goperl_pp_slot_332; \
    (tbl)[333] = goperl_pp_slot_333; \
    (tbl)[334] = goperl_pp_slot_334; \
    (tbl)[335] = goperl_pp_slot_335; \
    (tbl)[336] = goperl_pp_slot_336; \
    (tbl)[337] = goperl_pp_slot_337; \
    (tbl)[338] = goperl_pp_slot_338; \
    (tbl)[339] = goperl_pp_slot_339; \
    (tbl)[340] = goperl_pp_slot_340; \
    (tbl)[341] = goperl_pp_slot_341; \
    (tbl)[342] = goperl_pp_slot_342; \
    (tbl)[343] = goperl_pp_slot_343; \
    (tbl)[344] = goperl_pp_slot_344; \
    (tbl)[345] = goperl_pp_slot_345; \
    (tbl)[346] = goperl_pp_slot_346; \
    (tbl)[347] = goperl_pp_slot_347; \
    (tbl)[348] = goperl_pp_slot_348; \
    (tbl)[349] = goperl_pp_slot_349; \
    (tbl)[350] = goperl_pp_slot_350; \
    (tbl)[351] = goperl_pp_slot_351; \
    (tbl)[352] = goperl_pp_slot_352; \
    (tbl)[353] = goperl_pp_slot_353; \
    (tbl)[354] = goperl_pp_slot_354; \
    (tbl)[355] = goperl_pp_slot_355; \
    (tbl)[356] = goperl_pp_slot_356; \
    (tbl)[357] = goperl_pp_slot_357; \
    (tbl)[358] = goperl_pp_slot_358; \
    (tbl)[359] = goperl_pp_slot_359; \
    (tbl)[360] = goperl_pp_slot_360; \
    (tbl)[361] = goperl_pp_slot_361; \
    (tbl)[362] = goperl_pp_slot_362; \
    (tbl)[363] = goperl_pp_slot_363; \
    (tbl)[364] = goperl_pp_slot_364; \
    (tbl)[365] = goperl_pp_slot_365; \
    (tbl)[366] = goperl_pp_slot_366; \
    (tbl)[367] = goperl_pp_slot_367; \
    (tbl)[368] = goperl_pp_slot_368; \
    (tbl)[369] = goperl_pp_slot_369; \
    (tbl)[370] = goperl_pp_slot_370; \
    (tbl)[371] = goperl_pp_slot_371; \
    (tbl)[372] = goperl_pp_slot_372; \
    (tbl)[373] = goperl_pp_slot_373; \
    (tbl)[374] = goperl_pp_slot_374; \
    (tbl)[375] = goperl_pp_slot_375; \
    (tbl)[376] = goperl_pp_slot_376; \
    (tbl)[377] = goperl_pp_slot_377; \
    (tbl)[378] = goperl_pp_slot_378; \
    (tbl)[379] = goperl_pp_slot_379; \
    (tbl)[380] = goperl_pp_slot_380; \
    (tbl)[381] = goperl_pp_slot_381; \
    (tbl)[382] = goperl_pp_slot_382; \
    (tbl)[383] = goperl_pp_slot_383; \
    (tbl)[384] = goperl_pp_slot_384; \
    (tbl)[385] = goperl_pp_slot_385; \
    (tbl)[386] = goperl_pp_slot_386; \
    (tbl)[387] = goperl_pp_slot_387; \
    (tbl)[388] = goperl_pp_slot_388; \
    (tbl)[389] = goperl_pp_slot_389; \
    (tbl)[390] = goperl_pp_slot_390; \
    (tbl)[391] = goperl_pp_slot_391; \
    (tbl)[392] = goperl_pp_slot_392; \
    (tbl)[393] = goperl_pp_slot_393; \
    (tbl)[394] = goperl_pp_slot_394; \
    (tbl)[395] = goperl_pp_slot_395; \
    (tbl)[396] = goperl_pp_slot_396; \
    (tbl)[397] = goperl_pp_slot_397; \
    (tbl)[398] = goperl_pp_slot_398; \
    (tbl)[399] = goperl_pp_slot_399; \
    (tbl)[400] = goperl_pp_slot_400; \
    (tbl)[401] = goperl_pp_slot_401; \
    (tbl)[402] = goperl_pp_slot_402; \
    (tbl)[403] = goperl_pp_slot_403; \
    (tbl)[404] = goperl_pp_slot_404; \
    (tbl)[405] = goperl_pp_slot_405; \
    (tbl)[406] = goperl_pp_slot_406; \
    (tbl)[407] = goperl_pp_slot_407; \
    (tbl)[408] = goperl_pp_slot_408; \
    (tbl)[409] = goperl_pp_slot_409; \
    (tbl)[410] = goperl_pp_slot_410; \
    (tbl)[411] = goperl_pp_slot_411; \
    (tbl)[412] = goperl_pp_slot_412; \
    (tbl)[413] = goperl_pp_slot_413; \
    (tbl)[414] = goperl_pp_slot_414; \
    (tbl)[415] = goperl_pp_slot_415; \
    (tbl)[416] = goperl_pp_slot_416; \
    (tbl)[417] = goperl_pp_slot_417; \
    (tbl)[418] = goperl_pp_slot_418; \
    (tbl)[419] = goperl_pp_slot_419; \
    (tbl)[420] = goperl_pp_slot_420; \
    (tbl)[421] = goperl_pp_slot_421; \
    (tbl)[422] = goperl_pp_slot_422; \
    (tbl)[423] = goperl_pp_slot_423; \
    (tbl)[424] = goperl_pp_slot_424; \
    (tbl)[425] = goperl_pp_slot_425; \
} while (0)

#endif /* GOPERL_PPSLOTS_H */
